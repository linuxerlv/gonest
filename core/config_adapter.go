package core

import (
	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
)

type ConfigAdapter struct {
	cfg config.Config
}

func NewConfigAdapter(cfg config.Config) *ConfigAdapter {
	return &ConfigAdapter{cfg: cfg}
}

func (c *ConfigAdapter) GetString(key string) string {
	if c.cfg == nil {
		return ""
	}
	return c.cfg.GetString(key)
}

func (c *ConfigAdapter) GetInt(key string) int {
	if c.cfg == nil {
		return 0
	}
	return c.cfg.GetInt(key)
}

func (c *ConfigAdapter) GetBool(key string) bool {
	if c.cfg == nil {
		return false
	}
	return c.cfg.GetBool(key)
}

func (c *ConfigAdapter) Get(key string) any {
	if c.cfg == nil {
		return nil
	}
	return c.cfg.Get(key)
}

func (c *ConfigAdapter) GetDefault(key string, defaultValue any) any {
	if c.cfg == nil {
		return defaultValue
	}
	if !c.cfg.IsSet(key) {
		return defaultValue
	}
	return c.cfg.Get(key)
}

func (c *ConfigAdapter) Unmarshal(key string, v any) error {
	if c.cfg == nil {
		return nil
	}
	return c.cfg.Unmarshal(key, v)
}

func (c *ConfigAdapter) Unwrap() config.Config {
	return c.cfg
}

var _ abstract.Config = (*ConfigAdapter)(nil)
