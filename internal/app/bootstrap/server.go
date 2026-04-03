package bootstrap

import (
	"context"
	"fmt"

	"github.com/dovetaill/PureMux/pkg/config"
)

// BuildServerRuntime 组装 server 入口运行所需的共享资源。
func BuildServerRuntime(configPath string) (*Runtime, error) {
	rt, err := buildRuntime(configPath)
	if err != nil {
		return nil, err
	}

	if err := bootstrapServerBusinessSchema(rt); err != nil {
		_ = rt.Shutdown()
		return nil, fmt.Errorf("bootstrap server business schema: %w", err)
	}

	return rt, nil
}

func bootstrapServerBusinessSchema(rt *Runtime) error {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}

	if err := autoMigrateBusinessTablesFn(rt.Resources.MySQL); err != nil {
		return err
	}

	store := newSeedAdminStoreFn(rt.Resources)
	if store == nil {
		return nil
	}

	var seedCfg config.SeedAdminConfig
	if rt.Config != nil {
		seedCfg = rt.Config.Auth.SeedAdmin
	}

	return SeedDefaultAdmin(context.Background(), seedCfg, store, seedAdminPasswordHashFn)
}
