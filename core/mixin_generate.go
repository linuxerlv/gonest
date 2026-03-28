//go:build !wireinject
// +build !wireinject

package core

//go:generate go run ../internal/mixingen/cmd/main.go

// Mixin 代码生成指令
// 运行 go generate ./... 来生成 Mixin 代码
//
// Mixin 定义:
// - MiddlewareMixin: 中间件扩展方法
//
// 目标类型:
// - WebApplication: Web 应用
//
// 生成的方法:
// - UseCORS()
// - UseRecovery()
// - UseLogging()
// - UseRateLimit()
// - UseGzip()
// - UseSecurity()
// - UseRequestID()
// - UseTimeout()
