package bucketcontroller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// BucketValidator validates admission requests.
type BucketValidator struct {
	log logr.Logger
}

// ValidateCreate implements admission.CustomValidator.
func (v *BucketValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	res := obj.(*virtuozzov1.Bucket)
	v.log.V(1).Info("Validate create (noop)", "name", res.Name)
	return nil, nil
}

// ValidateUpdate implements admission.CustomValidator.
func (v *BucketValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	newBucket := newObj.(*virtuozzov1.Bucket)
	oldBucket := oldObj.(*virtuozzov1.Bucket)
	v.log.V(1).Info("Validate update")

	if oldBucket.Status.AtProvider.BucketName != "" {
		if newBucket.GetBucketName() != oldBucket.Status.AtProvider.BucketName {
			return nil, fmt.Errorf("a bucket named %q has been created already, you cannot rename it",
				oldBucket.Status.AtProvider.BucketName)
		}
	}
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator.
func (v *BucketValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	res := obj.(*virtuozzov1.Bucket)
	v.log.V(1).Info("Validate delete (noop)", "name", res.Name)
	return nil, nil
}
