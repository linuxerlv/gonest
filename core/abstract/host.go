package abstract

// Host 主机接口
type Host interface {
	CreateBuilder(args ...string) WebApplicationBuilder
}

// HostApplicationBuilder 主机应用构建器接口
type HostApplicationBuilder interface {
	ConfigureServices(configure func(ServiceCollection)) HostApplicationBuilder
	Configure(configure func(WebApplicationBuilder)) HostApplicationBuilder
	Build() WebApplication
}
