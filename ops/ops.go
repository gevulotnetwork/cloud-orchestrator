package ops

import (
	api "github.com/nanovms/ops/lepton"
	"github.com/nanovms/ops/provider"
	"github.com/nanovms/ops/types"
)

func Provider(cfg *types.Config) (api.Provider, *api.Context, error) {
	p, err := provider.CloudProvider(cfg.CloudConfig.Platform, &cfg.CloudConfig)
	if err != nil {
		return nil, nil, err
	}

	ctx := api.NewContext(cfg)

	return p, ctx, nil
}
