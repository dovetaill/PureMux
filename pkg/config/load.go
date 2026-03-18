package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

// Load 从 YAML 文件读取配置，并允许环境变量覆盖字段值。
func Load(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("config path is required")
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	return cfg, nil
}
