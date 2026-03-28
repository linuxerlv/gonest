package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gonest new <project-name>")
			return
		}
		createProject(os.Args[2])
	case "controller":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gonest controller <name>")
			return
		}
		generateController(os.Args[2])
	case "service":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gonest service <name>")
			return
		}
		generateService(os.Args[2])
	case "middleware":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gonest middleware <name>")
			return
		}
		generateMiddleware(os.Args[2])
	case "guard":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gonest guard <name>")
			return
		}
		generateGuard(os.Args[2])
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`Gonest CLI - Code Generator

Usage:
  gonest new <project>        Create a new project
  gonest controller <name>    Generate a controller
  gonest service <name>       Generate a service
  gonest middleware <name>    Generate a middleware
  gonest guard <name>         Generate a guard

Examples:
  gonest new myapp
  gonest controller user
  gonest service auth
  gonest middleware logging`)
}

func createProject(name string) {
	dirs := []string{
		name,
		name + "/cmd",
		name + "/internal/controllers",
		name + "/internal/services",
		name + "/internal/middleware",
		name + "/internal/models",
		name + "/wire",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	writeFile(name+"/go.mod", goModTemplate, map[string]string{"Name": name})
	writeFile(name+"/cmd/main.go", mainTemplate, map[string]string{"Name": name})
	writeFile(name+"/wire/wire.go", wireTemplate, map[string]string{"Name": name})
	writeFile(name+"/internal/controllers/app.controller.go", appControllerTemplate, nil)
	writeFile(name+"/internal/middleware/logger.middleware.go", loggerMiddlewareTemplate, nil)
	writeFile(name+"/internal/services/user.service.go", userServiceTemplate, nil)

	fmt.Printf("\n✅ Project '%s' created successfully!\n\n", name)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  go mod tidy")
	fmt.Println("  go generate ./wire/...")
	fmt.Println("  go run cmd/main.go")
}

func generateController(name string) {
	dir := "internal/controllers"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = "."
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".controller.go")
	data := map[string]string{
		"Name":  toPascalCase(name),
		"name":  strings.ToLower(name),
		"Name_": strings.ToLower(name),
	}

	writeFile(filename, controllerTemplate, data)
	fmt.Printf("✅ Controller '%s' created: %s\n", name, filename)
}

func generateService(name string) {
	dir := "internal/services"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = "."
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".service.go")
	data := map[string]string{
		"Name": toPascalCase(name),
		"name": strings.ToLower(name),
	}

	writeFile(filename, serviceTemplate, data)
	fmt.Printf("✅ Service '%s' created: %s\n", name, filename)
}

func generateMiddleware(name string) {
	dir := "internal/middleware"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = "."
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".middleware.go")
	data := map[string]string{
		"Name": toPascalCase(name),
		"name": strings.ToLower(name),
	}

	writeFile(filename, middlewareTemplate, data)
	fmt.Printf("✅ Middleware '%s' created: %s\n", name, filename)
}

func generateGuard(name string) {
	dir := "internal/guards"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = "."
	}

	filename := filepath.Join(dir, strings.ToLower(name)+".guard.go")
	data := map[string]string{
		"Name": toPascalCase(name),
		"name": strings.ToLower(name),
	}

	writeFile(filename, guardTemplate, data)
	fmt.Printf("✅ Guard '%s' created: %s\n", name, filename)
}

func writeFile(filename, content string, data map[string]string) {
	tmpl, err := template.New("file").Parse(content)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	if data == nil {
		data = make(map[string]string)
	}

	if err := tmpl.Execute(file, data); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
	}
}

func toPascalCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "-")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return result
}

var goModTemplate = `module {{.Name}}

go 1.22

require (
	github.com/google/wire v0.6.0
	github.com/linuxerlv/gonest v0.0.0
)

replace github.com/linuxerlv/gonest => ../../gonest
`

var mainTemplate = `package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/core"
	"{{.Name}}/wire"
)

func main() {
	app := wire.InitializeApp()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.Shutdown(ctx)
		log.Println("Server stopped")
	}()

	log.Println("Server running on :8080")
	if err := app.Listen(":8080"); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
`

var wireTemplate = `//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/linuxerlv/gonest"
	"github.com/linuxerlv/gonest/core"
	"{{.Name}}/internal/controllers"
	"{{.Name}}/internal/middleware"
	"{{.Name}}/internal/services"
)

func InitializeApp() *core.WebApplication {
	wire.Build(
		core.ProvideWebApplicationBuilder,
		core.ProvideServiceCollection,
		ProvideWebApplication,
		controllers.NewAppController,
		middleware.NewLoggerMiddleware,
		services.NewUserService,
	)
	return nil
}

func ProvideWebApplication(
	builder *core.WebApplicationBuilder,
	appController *controllers.AppController,
	loggerMiddleware gonest.Middleware,
) *core.WebApplication {
	app := builder.BuildWeb()

	app.Use(loggerMiddleware)

	app.MapGet("/", appController.Index)
	app.MapGet("/health", appController.Health)

	return app
}
`

var appControllerTemplate = `package controllers

import (
	"net/http"

	"github.com/linuxerlv/gonest"
)

type AppController struct{}

func NewAppController() *AppController {
	return &AppController{}
}

func (c *AppController) Index(ctx gonest.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Welcome to Gonest",
		"version": "1.0.0",
	})
}

func (c *AppController) Health(ctx gonest.Context) error {
	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
`

var loggerMiddlewareTemplate = `package middleware

import (
	"log"
	"time"

	"github.com/linuxerlv/gonest"
)

func NewLoggerMiddleware() gonest.Middleware {
	return gonest.MiddlewareFunc(func(ctx gonest.Context, next func() error) error {
		start := time.Now()
		err := next()
		log.Printf("[HTTP] %s %s %v", ctx.Method(), ctx.Path(), time.Since(start))
		return err
	})
}
`

var userServiceTemplate = `package services

type UserService interface {
	GetByID(id string) (any, error)
	GetAll() ([]any, error)
}

type UserServiceImpl struct{}

func NewUserService() UserService {
	return &UserServiceImpl{}
}

func (s *UserServiceImpl) GetByID(id string) (any, error) {
	return map[string]string{"id": id, "name": "User"}, nil
}

func (s *UserServiceImpl) GetAll() ([]any, error) {
	return []any{
		map[string]string{"id": "1", "name": "User 1"},
		map[string]string{"id": "2", "name": "User 2"},
	}, nil
}
`

var controllerTemplate = `package controllers

import (
	"net/http"

	"github.com/linuxerlv/gonest"
)

type {{.Name}}Controller struct{}

func New{{.Name}}Controller() *{{.Name}}Controller {
	return &{{.Name}}Controller{}
}

func (c *{{.Name}}Controller) List(ctx gonest.Context) error {
	return ctx.JSON(http.StatusOK, map[string]any{
		"data": []any{},
	})
}

func (c *{{.Name}}Controller) Get(ctx gonest.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(http.StatusOK, map[string]string{
		"id": id,
	})
}

func (c *{{.Name}}Controller) Create(ctx gonest.Context) error {
	var input struct {
		Name string ` + "`json:\"name\"`" + `
	}
	if err := ctx.Bind(&input); err != nil {
		return gonest.BadRequest("invalid input")
	}
	return ctx.JSON(http.StatusCreated, input)
}

func (c *{{.Name}}Controller) Update(ctx gonest.Context) error {
	id := ctx.Param("id")
	var input struct {
		Name string ` + "`json:\"name\"`" + `
	}
	if err := ctx.Bind(&input); err != nil {
		return gonest.BadRequest("invalid input")
	}
	return ctx.JSON(http.StatusOK, map[string]string{
		"id":   id,
		"name": input.Name,
	})
}

func (c *{{.Name}}Controller) Delete(ctx gonest.Context) error {
	id := ctx.Param("id")
	return ctx.JSON(http.StatusOK, map[string]string{
		"id":      id,
		"deleted": "true",
	})
}
`

var serviceTemplate = `package services

import "github.com/linuxerlv/gonest/core"

type {{.Name}}Service interface {
	FindAll() ([]any, error)
	FindByID(id string) (any, error)
	Create(data any) (any, error)
	Update(id string, data any) (any, error)
	Delete(id string) error
}

type {{.Name}}ServiceImpl struct {
	services *core.ServiceCollection
}

func New{{.Name}}Service(services *core.ServiceCollection) {{.Name}}Service {
	return &{{.Name}}ServiceImpl{services: services}
}

func (s *{{.Name}}ServiceImpl) FindAll() ([]any, error) {
	return []any{}, nil
}

func (s *{{.Name}}ServiceImpl) FindByID(id string) (any, error) {
	return map[string]string{"id": id}, nil
}

func (s *{{.Name}}ServiceImpl) Create(data any) (any, error) {
	return data, nil
}

func (s *{{.Name}}ServiceImpl) Update(id string, data any) (any, error) {
	return map[string]any{"id": id, "data": data}, nil
}

func (s *{{.Name}}ServiceImpl) Delete(id string) error {
	return nil
}
`

var middlewareTemplate = `package middleware

import (
	"github.com/linuxerlv/gonest"
)

func New{{.Name}}Middleware() gonest.Middleware {
	return gonest.MiddlewareFunc(func(ctx gonest.Context, next func() error) error {
		return next()
	})
}
`

var guardTemplate = `package guards

import (
	"github.com/linuxerlv/gonest"
)

func New{{.Name}}Guard() gonest.Guard {
	return gonest.GuardFunc(func(ctx gonest.Context) bool {
		return true
	})
}
`
