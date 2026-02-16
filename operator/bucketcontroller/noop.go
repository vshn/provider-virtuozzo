package bucketcontroller

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// NoopClient is a client that does nothing.
type NoopClient struct{}

func (n *NoopClient) Disconnect(ctx context.Context) error {
	return nil
}

// Observe implements managed.ExternalClient.
func (n *NoopClient) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}

// Create implements managed.ExternalClient.
func (n *NoopClient) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}

// Update implements managed.ExternalClient.
func (n *NoopClient) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete implements managed.ExternalClient.
func (n *NoopClient) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}
