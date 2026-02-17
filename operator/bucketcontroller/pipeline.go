package bucketcontroller

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/minio/minio-go/v7"
	providerv1 "github.com/vshn/provider-virtuozzo/apis/provider/v1"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
)

// ProvisioningPipeline provisions Buckets using S3 client.
type ProvisioningPipeline struct {
	recorder    event.Recorder
	minio       *minio.Client
	endpointURL string
	accessKey   string
	secretKey   string
}

func (p *ProvisioningPipeline) Disconnect(ctx context.Context) error {
	return nil
}

type pipelineContext struct {
	context.Context
	bucket *virtuozzov1.Bucket
}

// NewProvisioningPipeline returns a new instance of ProvisioningPipeline.
func NewProvisioningPipeline(recorder event.Recorder, minio *minio.Client, endpointURL, accessKey, secretKey string) *ProvisioningPipeline {
	return &ProvisioningPipeline{
		recorder:    recorder,
		minio:       minio,
		endpointURL: endpointURL,
		accessKey:   accessKey,
		secretKey:   secretKey,
	}
}

func fromManaged(mg resource.Managed) *virtuozzov1.Bucket {
	return mg.(*virtuozzov1.Bucket)
}

const lockAnnotation = virtuozzov1.Group + "/lock"

func (p *ProvisioningPipeline) connectionDetails() managed.ConnectionDetails {
	return managed.ConnectionDetails{
		providerv1.AccessKeyIDKey:     []byte(p.accessKey),
		providerv1.SecretAccessKeyKey: []byte(p.secretKey),
	}
}
