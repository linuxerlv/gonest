package config

import (
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

var _ abstract.Config = (*KoanfConfig)(nil)

type Config interface {
	Get(key string) any
	GetString(key string) string
	GetInt(key string) int
	GetInt64(key string) int64
	GetFloat64(key string) float64
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetTime(key string) time.Time
	GetStringSlice(key string) []string
	GetIntSlice(key string) []int
	GetStringMap(key string) map[string]any
	GetStringMapString(key string) map[string]string

	IsSet(key string) bool
	Unmarshal(key string, out any) error
	UnmarshalWithConf(key string, out any, opts UnmarshalOptions) error

	Load(provider Provider, parser Parser) error
	LoadWithOverride(provider Provider, parser Parser, override func(map[string]any)) error

	Merge(other Config) error

	All() map[string]any
	Keys() []string

	Print()
}

type UnmarshalOptions struct {
	Tag       string
	FlatPaths bool
}

type Provider interface {
	Read() (map[string]any, error)
	Watch(callback func(any, error)) error
	Name() string
}

type Parser interface {
	Parse(data []byte) (map[string]any, error)
	Marshal(data map[string]any) ([]byte, error)
	Name() string
}

type ConfigSource struct {
	Name     string
	Priority int
	Provider Provider
	Parser   Parser
}

type ConfigBuilder interface {
	AddSource(source ConfigSource) ConfigBuilder
	AddFile(path string, parser Parser) ConfigBuilder
	AddEnv(prefix string) ConfigBuilder
	AddDefaults(defaults map[string]any) ConfigBuilder
	Build() (Config, error)
}

type WatchEvent struct {
	Key       string
	OldValue  any
	NewValue  any
	Timestamp time.Time
}

func Get[T any](c Config, key string) T {
	v := c.Get(key)
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func GetDefault[T any](c Config, key string, defaultValue T) T {
	if !c.IsSet(key) {
		return defaultValue
	}
	return Get[T](c, key)
}
