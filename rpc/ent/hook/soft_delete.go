package hook

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/intercept"
	"github.com/coder-lulu/newbee-core/rpc/ent/schema/mixins"
)

type softDeleteKey struct{}

// SkipSoftDelete returns a new context that skips the soft-delete interceptor.
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// SoftDeleteInterceptor returns an interceptor that filters out soft-deleted entities.
func SoftDeleteInterceptor() ent.Interceptor {
	mixin := mixins.SoftDeleteMixin{}
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// Skip soft-delete, means include soft-deleted entities.
		if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
			return nil
		}
		
		// Apply soft delete filter
		mixin.P(q)
		return nil
	})
}
