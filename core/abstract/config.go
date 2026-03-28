package abstract

// Config 配置接口
type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	Get(key string) any
	GetDefault(key string, defaultValue any) any
	Unmarshal(key string, v any) error
}

// TypedConfig 泛型配置接口
type TypedConfig interface {
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// ConfigLoader 配置加载接口
type ConfigLoader interface {
	Load() error
	LoadFile(path string) error
}

// ConfigWatcher 配置监听接口
type ConfigWatcher interface {
	Watch(callback func(key string, value any))
	StopWatch()
}

// ConfigProvider 配置提供者接口
type ConfigProvider interface {
	Name() string
	Provide() (map[string]any, error)
}

// Configurable 可配置接口
type Configurable interface {
	Configure(cfg Config) error
}
