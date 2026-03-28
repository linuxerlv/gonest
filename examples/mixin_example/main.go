package main

import (
	"fmt"
	"net/http"

	"github.com/linuxerlv/gonest/core"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/extensions"
)

func main() {
	builder := core.NewWebApplicationBuilder()

	builder.Services().AddCORS(&extensions.CORSMiddlewareOptions{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST"},
	})

	builder.Services().AddRecovery(nil)
	builder.Services().AddLogging(nil)

	app := core.NewWebAppWithMixin(
		builder.Build().(*core.WebApplication),
		builder.Services().(*core.ServiceCollection),
	)

	app.UseCORS().
		UseRecovery().
		UseLogging()

	app.MapGet("/", func(ctx abstract.Context) error {
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Hello, Mixin + Wire!",
		})
	})

	fmt.Println("Server starting on :8080")
	app.Listen(":8080")
}
