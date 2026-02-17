// Package apis contains Kubernetes API for the Virtuozzo provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	providerv1 "github.com/vshn/provider-virtuozzo/apis/provider/v1"
	virtuozzov1 "github.com/vshn/provider-virtuozzo/apis/virtuozzo/v1"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		virtuozzov1.SchemeBuilder.AddToScheme,
		providerv1.SchemeBuilder.AddToScheme,
	)
}

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
