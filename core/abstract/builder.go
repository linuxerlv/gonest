package abstract

// WebApplicationBuilderAbstract Web应用构建器接口
type WebApplicationBuilderAbstract interface {
	UseConfig(cfg ConfigAbstract) WebApplicationBuilderAbstract
	UseLogger(log LoggerAbstract) WebApplicationBuilderAbstract
	ConfigureServices(configure func(ServiceCollectionAbstract)) WebApplicationBuilderAbstract
	Build() WebApplicationAbstract
}

// WebApplicationAbstract Web应用接口
type WebApplicationAbstract interface {
	ApplicationAbstract
	Configuration() ConfigAbstract
	Log() LoggerAbstract
	Run() error
	RunAsync() <-chan error
	WaitForShutdown() error
}

// MapRouteAbstract 路由映射接口
type MapRouteAbstract interface {
	MapGet(path string, handler any) *RouteBuilderAbstract
	MapPost(path string, handler any) *RouteBuilderAbstract
	MapPut(path string, handler any) *RouteBuilderAbstract
	MapDelete(path string, handler any) *RouteBuilderAbstract
	MapPatch(path string, handler any) *RouteBuilderAbstract
}
