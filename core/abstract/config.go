package abstract

// ConfigAbstract 配置接口
type ConfigAbstract interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Get(key string) any
	GetDefault(key string, defaultValue any) any
	Unmarshal(key string, v any) error
}

// TypedConfigAbstract 泛型配置接口
type TypedConfigAbstract interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// ConfigLoaderAbstract 配置加载接口
type ConfigLoaderAbstract interface {
	Load() error
	LoadFile(path string) error
}

// ConfigWatcherAbstract 配置监听接口
type ConfigWatcherAbstract interface {
	Watch(callback func(key string, value any))
	StopWatch()
}

// ConfigProviderAbstract 配置提供者接口
type ConfigProviderAbstract interface {
	Name() string
	Provide() (map[string]any, error)
}

// ConfigurableAbstract 可配置接口
type ConfigurableAbstract interface {
	Configure(cfg ConfigAbstract) error
}
