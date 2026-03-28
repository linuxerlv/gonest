package gonest

import (
	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
)

type Context = abstract.ContextAbstract
type HttpContext = core.HttpContext

type Router = abstract.RouterAbstract
type HttpRouter = core.HttpRouter

type Application = abstract.ApplicationAbstract
type WebApplication = core.WebApplication
type WebApplicationBuilder = core.WebApplicationBuilder

type Middleware = abstract.MiddlewareAbstract
type MiddlewareFunc = abstract.MiddlewareFuncAbstract

type RouteHandler = abstract.RouteHandlerAbstract
type RouteBuilder = abstract.RouteBuilderAbstract
type RouteGroup = abstract.RouteGroupAbstract
type Route = core.Route

type Guard = abstract.GuardAbstract
type GuardFunc = abstract.GuardFuncAbstract

type Interceptor = abstract.InterceptorAbstract
type InterceptorFunc = abstract.InterceptorFuncAbstract

type Pipe = abstract.PipeAbstract
type PipeFunc = abstract.PipeFuncAbstract

type ExceptionFilter = abstract.ExceptionFilterAbstract

type Controller = abstract.ControllerAbstract

type ServiceCollection = core.ServiceCollection
type ServiceDescriptor = core.ServiceDescriptor

type HttpException = abstract.HttpException

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
var NewHostBuilder = core.NewHostBuilder

type CORSConfig = abstract.CORSConfigAbstract
type SecurityConfig = abstract.SecurityConfigAbstract
type LoggerMiddlewareConfig = abstract.LoggerMiddlewareConfigAbstract

var _ abstract.ContextAbstract = (*core.HttpContext)(nil)
