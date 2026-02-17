package bucketcontroller

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	pipeline "github.com/ccremer/go-command-pipeline"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	providerv1 "github.com/vshn/provider-virtuozzo/apis/provider/v1"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
	"github.com/vshn/provider-virtuozzo/operator/pipelineutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type bucketConnector struct {
	kube     client.Client
	recorder event.Recorder
}

type connectContext struct {
	context.Context
	bucket            *virtuozzov1.Bucket
	providerConfig    *providerv1.ProviderConfig
	credentialsSecret *corev1.Secret
	minio             *minio.Client
	endpointURL       string
	accessKey         string
	secretKey         string
}

// Connect implements managed.ExternalConnector.
func (c *bucketConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	ctx = pipeline.MutableContext(ctx)
	log := controllerruntime.LoggerFrom(ctx)
	log.V(1).Info("Connecting resource")

	bucket := fromManaged(mg)

	if isBucketAlreadyDeleted(bucket) {
		log.V(1).Info("Bucket already deleted, skipping observation")
		return &NoopClient{}, nil
	}

	pctx := &connectContext{Context: ctx, bucket: bucket}
	pipe := pipeline.NewPipeline[*connectContext]()
	pipe.WithBeforeHooks(pipelineutil.DebugLogger(pctx)).
		WithSteps(
			pipe.NewStep("track provider config", c.trackProviderConfig),
			pipe.NewStep("fetch provider config", c.fetchProviderConfig),
			pipe.NewStep("fetch credentials secret", c.fetchCredentialsSecret),
			pipe.NewStep("read credentials", c.readCredentials),
			pipe.NewStep("create S3 client", c.createS3Client),
		)
	result := pipe.RunWithContext(pctx)

	if result != nil {
		return nil, result
	}

	return NewProvisioningPipeline(c.recorder, pctx.minio, pctx.endpointURL, pctx.accessKey, pctx.secretKey), nil
}

func (c *bucketConnector) trackProviderConfig(ctx *connectContext) error {
	return resource.NewProviderConfigUsageTracker(c.kube, &providerv1.ProviderConfigUsage{}).Track(ctx, ctx.bucket)
}

func (c *bucketConnector) fetchProviderConfig(ctx *connectContext) error {
	pcRef := ctx.bucket.GetProviderConfigReference()
	if pcRef == nil {
		return fmt.Errorf("providerConfigRef is not set on the resource")
	}

	pc := &providerv1.ProviderConfig{}
	err := c.kube.Get(ctx, types.NamespacedName{Name: pcRef.Name}, pc)
	if err != nil {
		return fmt.Errorf("cannot get ProviderConfig %q: %w", pcRef.Name, err)
	}
	ctx.providerConfig = pc
	return nil
}

func (c *bucketConnector) fetchCredentialsSecret(ctx *connectContext) error {
	secretRef := ctx.providerConfig.Spec.Credentials.CredentialsSecretRef
	if secretRef.Name == "" {
		return fmt.Errorf("credentialsSecretRef.name is not set in ProviderConfig %q", ctx.providerConfig.Name)
	}

	secret := &corev1.Secret{}
	err := c.kube.Get(ctx, types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}, secret)
	if err != nil {
		return fmt.Errorf("cannot get credentials Secret %q: %w", fmt.Sprintf("%s/%s", secretRef.Namespace, secretRef.Name), err)
	}
	ctx.credentialsSecret = secret
	return nil
}

func (c *bucketConnector) readCredentials(ctx *connectContext) error {
	secret := ctx.credentialsSecret
	secretName := fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)

	if secret.Data == nil {
		return fmt.Errorf("secret %q does not have any data", secretName)
	}

	accessKey := string(secret.Data[providerv1.AccessKeyIDKey])
	secretKey := string(secret.Data[providerv1.SecretAccessKeyKey])
	endpointURL := string(secret.Data[providerv1.EndpointURLKey])

	if accessKey == "" {
		return fmt.Errorf("secret %q is missing key %s or its value is empty", secretName, providerv1.AccessKeyIDKey)
	}
	if secretKey == "" {
		return fmt.Errorf("secret %q is missing key %s or its value is empty", secretName, providerv1.SecretAccessKeyKey)
	}
	if endpointURL == "" {
		return fmt.Errorf("secret %q is missing key %s or its value is empty", secretName, providerv1.EndpointURLKey)
	}

	ctx.accessKey = accessKey
	ctx.secretKey = secretKey
	ctx.endpointURL = endpointURL
	return nil
}

// createS3Client creates a new client using the S3 credentials from the Secret.
func (c *bucketConnector) createS3Client(ctx *connectContext) error {
	parsed, err := url.Parse(ctx.endpointURL)
	if err != nil {
		return fmt.Errorf("cannot parse endpoint URL %q: %w", ctx.endpointURL, err)
	}

	host := parsed.Host
	if parsed.Host == "" {
		host = parsed.Path // if no scheme is given, it's parsed as a path
	}
	s3Client, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(ctx.accessKey, ctx.secretKey, ""),
		Secure: isTLSEnabled(parsed),
	})
	ctx.minio = s3Client
	return err
}

// isBucketAlreadyDeleted returns true if the status conditions are in a state where one can assume that the deletion of a bucket was successful in a previous reconciliation.
func isBucketAlreadyDeleted(bucket *virtuozzov1.Bucket) bool {
	readyCond := findCondition(bucket.Status.Conditions, xpv1.TypeReady)
	syncCond := findCondition(bucket.Status.Conditions, xpv1.TypeSynced)

	if readyCond != nil && syncCond != nil {
		if readyCond.Status == corev1.ConditionFalse &&
			readyCond.Reason == xpv1.ReasonDeleting &&
			syncCond.Status == corev1.ConditionTrue &&
			syncCond.Reason == xpv1.ReasonReconcileSuccess {
			return true
		}
	}
	return false
}

func findCondition(conds []xpv1.Condition, condType xpv1.ConditionType) *xpv1.Condition {
	for _, cond := range conds {
		if cond.Type == condType {
			return &cond
		}
	}
	return nil
}

// isTLSEnabled returns false if the scheme is explicitly set to `http` or `HTTP`
func isTLSEnabled(u *url.URL) bool {
	return !strings.EqualFold(u.Scheme, "http")
}
