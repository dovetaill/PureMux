package bootstrap

import (
	"fmt"
)

// BuildServerRuntime 组装 server 入口运行所需的共享资源。
func BuildServerRuntime(configPath string) (*Runtime, error) {
	rt, err := buildRuntime(configPath)
	if err != nil {
		return nil, err
	}

	if err := bootstrapServerStarterSchema(rt); err != nil {
		_ = rt.Shutdown()
		return nil, fmt.Errorf("bootstrap server starter schema: %w", err)
	}

	return rt, nil
}

func bootstrapServerStarterSchema(rt *Runtime) error {
	if rt == nil || rt.Resources == nil || rt.Resources.MySQL == nil {
		return nil
	}

	return autoMigrateBusinessTablesFn(rt.Resources.MySQL)
}
