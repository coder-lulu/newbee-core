package oauthprovider

import (
	"context"
	"testing"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
)

func TestCreateProviderInitializesLegacySecretWithoutPlaintext(t *testing.T) {
	for _, withEmptySecret := range []bool{false, true} {
		t.Run(map[bool]string{false: "omitted", true: "empty"}[withEmptySecret], func(t *testing.T) {
			db := ent.NewClient()
			called := false
			db.OauthProvider.Use(func(next ent.Mutator) ent.Mutator {
				return ent.MutateFunc(func(ctx context.Context, mutation ent.Mutation) (ent.Value, error) {
					value, ok := mutation.(*ent.OauthProviderMutation).ClientSecret()
					if !ok || value != "" {
						t.Fatalf("legacy required column must be explicitly empty, present=%v", ok)
					}
					called = true
					return &ent.OauthProvider{ID: 1}, nil
				})
			})
			request := &core.OauthProviderInfo{}
			if withEmptySecret {
				empty := ""
				request.ClientSecret = &empty
			}
			response, err := NewCreateOauthProviderLogic(context.Background(), &svc.ServiceContext{DB: db}).CreateOauthProvider(request)
			if err != nil || !called || response.Id != 1 {
				t.Fatalf("create provider: response=%v err=%v called=%v", response, err, called)
			}
		})
	}
}
