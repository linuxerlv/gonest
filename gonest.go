package gonest

import (
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/extensions"
)

type Context = abstract.Context
type HttpContext = core.HttpContext

type Router = abstract.Router
type HttpRouter = core.HttpRouter

type Application = abstract.Application
type ApplicationBuilder = abstract.ApplicationBuilder
type WebApplication = abstract.WebApplication
type WebApplicationBuilder = abstract.WebApplicationBuilder

type HostApplication = core.HostApplication
type WebApp = core.WebApplication
type AppBuilder = core.ApplicationBuilder
type WebAppBuilder = core.WebApplicationBuilder

type Middleware = abstract.Middleware
type MiddlewareFunc = abstract.MiddlewareFunc

type RouteHandler = abstract.RouteHandler
type RouteBuilder = abstract.RouteBuilder
type RouteGroup = abstract.RouteGroup
type Route = core.Route

type Guard = abstract.Guard
type GuardFunc = abstract.GuardFunc

type Interceptor = abstract.Interceptor
type InterceptorFunc = abstract.InterceptorFunc

type Pipe = abstract.Pipe
type PipeFunc = abstract.PipeFunc

type ExceptionFilter = abstract.ExceptionFilter

type Controller = abstract.Controller

type ServiceCollection = core.ServiceCollection
type ServiceDescriptor = core.ServiceDescriptor

type HttpError = abstract.HttpError

type CORSMiddlewareOptions = extensions.CORSMiddlewareOptions
type RecoveryMiddlewareOptions = extensions.RecoveryMiddlewareOptions
type LoggingMiddlewareOptions = extensions.LoggingMiddlewareOptions
type RateLimitMiddlewareOptions = extensions.RateLimitMiddlewareOptions
type GzipMiddlewareOptions = extensions.GzipMiddlewareOptions
type SecurityMiddlewareOptions = extensions.SecurityMiddlewareOptions
type RequestIDMiddlewareOptions = extensions.RequestIDMiddlewareOptions
type TimeoutMiddlewareOptions = extensions.TimeoutMiddlewareOptions

var BadRequest = abstract.BadRequest
var Unauthorized = abstract.Unauthorized
var Forbidden = abstract.Forbidden
var NotFound = abstract.NotFound
var InternalError = abstract.InternalError
var NewHttpException = abstract.NewHttpException

var NewApplication = core.NewApplication
var NewRouter = core.NewRouter
var NewContext = core.NewContext
var NewContextWithParams = core.NewContextWithParams
var NewServiceCollection = core.NewServiceCollection
var CreateBuilder = core.CreateBuilder
var CreateApplication = core.CreateApplication
var NewWebApplicationBuilder = core.NewWebApplicationBuilder
var NewApplicationBuilder = core.NewApplicationBuilder
var NewHostBuilder = core.NewHostBuilder

var ApplicationCreateBuilder = core.ApplicationCreateBuilder
var WebApplicationCreateBuilder = core.WebApplicationCreateBuilder

var UseCORS = extensions.UseCORS
var UseRecovery = extensions.UseRecovery
var UseLogging = extensions.UseLogging
var UseRateLimit = extensions.UseRateLimit
var UseGzip = extensions.UseGzip
var UseSecurity = extensions.UseSecurity
var UseRequestID = extensions.UseRequestID
var UseTimeout = extensions.UseTimeout

var _ abstract.Context = (*core.HttpContext)(nil)
