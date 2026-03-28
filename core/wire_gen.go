//go:build wireinject
// +build wireinject

package core

import (
	"github.com/google/wire"
)

// WebAppWithMiddlewareSet Wire 集合，用于生成带有中间件扩展的 WebApplication
var WebAppWithMiddlewareSet = wire.NewSet(
	ProvideServiceCollection,
	ProvideWebApplicationBuilder,
	ProvideWebApplication,
	ProvideMiddlewareMixin,
	ProvideWebAppWithMixin,
)

// InitializeWebApp 初始化 WebApplication（带中间件扩展）
// Wire 会生成这个函数的实现
func InitializeWebApp() *WebAppWithMixin {
	wire.Build(WebAppWithMiddlewareSet)
	return nil
}

// InitializeWebAppWithServices 使用已有的 ServiceCollection 初始化
func InitializeWebAppWithServices(services *ServiceCollection) *WebAppWithMixin {
	wire.Build(
		ProvideWebApplicationBuilderFromServices,
		ProvideWebApplication,
		ProvideMiddlewareMixin,
		ProvideWebAppWithMixin,
	)
	return nil
}

// ProvideWebApplicationBuilderFromServices 从已有的 ServiceCollection 创建 Builder
func ProvideWebApplicationBuilderFromServices(services *ServiceCollection) *WebApplicationBuilder {
	builder := NewWebApplicationBuilder()
	builder.ApplicationBuilder.services = services
	return builder
}
