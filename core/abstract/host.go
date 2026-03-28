package abstract

// HostBuilderAbstract 主机构建器接口
type HostBuilderAbstract interface {
	UseContentRoot(path string) HostBuilderAbstract
	UseEnvironment(env string) HostBuilderAbstract
	ContentRoot() string
	Environment() string
	Args() []string
}

// HostAbstract 主机接口
type HostAbstract interface {
	CreateBuilder(args ...string) WebApplicationBuilderAbstract
}

// HostApplicationBuilderAbstract 主机应用构建器接口
type HostApplicationBuilderAbstract interface {
	ConfigureServices(configure func(ServiceCollectionAbstract)) HostApplicationBuilderAbstract
	Configure(configure func(WebApplicationBuilderAbstract)) HostApplicationBuilderAbstract
	Build() WebApplicationAbstract
}
