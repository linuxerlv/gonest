package config

import (
	"fmt"
	"time"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cast"
)

type KoanfConfig struct {
	k     *koanf.Koanf
	delim string
}

func NewKoanfConfig(delim string) *KoanfConfig {
	if delim == "" {
		delim = "."
	}
	return &KoanfConfig{
		k:     koanf.New(delim),
		delim: delim,
	}
}

func NewKoanfConfigWithConf(conf KoanfConf) *KoanfConfig {
	return &KoanfConfig{
		k:     koanf.NewWithConf(koanf.Conf(conf)),
		delim: conf.Delim,
	}
}

type KoanfConf struct {
	Delim       string
	StrictMerge bool
}

func DefaultKoanfConf() KoanfConf {
	return KoanfConf{
		Delim:       ".",
		StrictMerge: false,
	}
}

func (c *KoanfConfig) Load(provider Provider, parser Parser) error {
	data, err := provider.Read()
	if err != nil {
		return fmt.Errorf("failed to read config from provider %s: %w", provider.Name(), err)
	}

	if err := c.k.Load(confmap.Provider(data, c.delim), nil); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}
	return nil
}

func (c *KoanfConfig) LoadWithOverride(provider Provider, parser Parser, override func(map[string]any)) error {
	data, err := provider.Read()
	if err != nil {
		return fmt.Errorf("failed to read config from provider %s: %w", provider.Name(), err)
	}

	if override != nil {
		override(data)
	}

	if err := c.k.Load(confmap.Provider(data, c.delim), nil); err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}
	return nil
}

func (c *KoanfConfig) Merge(other Config) error {
	data := other.All()
	return c.k.Load(confmap.Provider(data, c.delim), nil)
}

func (c *KoanfConfig) Get(key string) any {
	return c.k.Get(key)
}

func (c *KoanfConfig) GetDefault(key string, defaultValue any) any {
	if !c.k.Exists(key) {
		return defaultValue
	}
	return c.k.Get(key)
}

func (c *KoanfConfig) GetString(key string) string {
	return c.k.String(key)
}

func (c *KoanfConfig) GetInt(key string) int {
	return c.k.Int(key)
}

func (c *KoanfConfig) GetInt64(key string) int64 {
	return c.k.Int64(key)
}

func (c *KoanfConfig) GetFloat64(key string) float64 {
	return c.k.Float64(key)
}

func (c *KoanfConfig) GetBool(key string) bool {
	return c.k.Bool(key)
}

func (c *KoanfConfig) GetDuration(key string) time.Duration {
	return c.k.Duration(key)
}

func (c *KoanfConfig) GetTime(key string) time.Time {
	return c.k.Time(key, time.RFC3339)
}

func (c *KoanfConfig) GetStringSlice(key string) []string {
	return c.k.Strings(key)
}

func (c *KoanfConfig) GetIntSlice(key string) []int {
	return c.k.Ints(key)
}

func (c *KoanfConfig) GetStringMap(key string) map[string]any {
	return cast.ToStringMap(c.k.Get(key))
}

func (c *KoanfConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(c.k.Get(key))
}

func (c *KoanfConfig) IsSet(key string) bool {
	return c.k.Exists(key)
}

func (c *KoanfConfig) Unmarshal(key string, out any) error {
	return c.k.Unmarshal(key, out)
}

func (c *KoanfConfig) UnmarshalWithConf(key string, out any, opts UnmarshalOptions) error {
	tag := opts.Tag
	if tag == "" {
		tag = "koanf"
	}
	return c.k.UnmarshalWithConf(key, out, koanf.UnmarshalConf{
		Tag:       tag,
		FlatPaths: opts.FlatPaths,
	})
}

func (c *KoanfConfig) All() map[string]any {
	return c.k.All()
}

func (c *KoanfConfig) Keys() []string {
	return c.k.Keys()
}

func (c *KoanfConfig) Print() {
	c.k.Print()
}

func (c *KoanfConfig) Raw() *koanf.Koanf {
	return c.k
}
