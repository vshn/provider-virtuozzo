package bucketcontroller

import (
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetupController adds a controller that reconciles virtuozzov1.Bucket managed resources.
func SetupController(mgr ctrl.Manager) error {
	name := managed.ControllerName(virtuozzov1.BucketGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	recorder := event.NewAPIRecorder(mgr.GetEventRecorderFor(name))

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(virtuozzov1.BucketGroupVersionKind),
		managed.WithExternalConnecter(&bucketConnector{
			kube:     mgr.GetClient(),
			recorder: recorder,
		}),
		managed.WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
		managed.WithRecorder(recorder),
		managed.WithPollInterval(1*time.Hour), // buckets are rather static
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&virtuozzov1.Bucket{}).
		Complete(r)
}

// SetupWebhook adds a webhook for Bucket managed resources.
func SetupWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&virtuozzov1.Bucket{}).
		WithValidator(&BucketValidator{
			log: mgr.GetLogger().WithName("webhook").WithName(strings.ToLower(virtuozzov1.BucketKind)),
		}).
		Complete()
}
