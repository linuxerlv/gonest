package core

import (
	"context"

	"github.com/linuxerlv/gonest/core/abstract"
)

type HostBuilder struct {
	contentRootPath string
	environment     string
	args            []string
	urls            []string
}

func NewHostBuilder() *HostBuilder {
	return &HostBuilder{
		environment: "development",
	}
}

func (h *HostBuilder) UseContentRoot(path string) abstract.HostBuilder {
	h.contentRootPath = path
	return h
}

func (h *HostBuilder) UseEnvironment(env string) abstract.HostBuilder {
	h.environment = env
	return h
}

func (h *HostBuilder) ContentRoot() string {
	return h.contentRootPath
}

func (h *HostBuilder) Environment() string {
	return h.environment
}

func (h *HostBuilder) Args() []string {
	return h.args
}

func (h *HostBuilder) SetArgs(args []string) {
	h.args = args
}

func (h *HostBuilder) UseUrls(urls ...string) abstract.WebHostBuilder {
	h.urls = urls
	return h
}

func (h *HostBuilder) ConfigureKestrel(configure func(interface{})) abstract.WebHostBuilder {
	return h
}

func (h *HostBuilder) Build() abstract.WebHost {
	return &WebHost{
		urls: h.urls,
	}
}

var _ abstract.HostBuilder = (*HostBuilder)(nil)
var _ abstract.WebHostBuilder = (*HostBuilder)(nil)

type WebHost struct {
	urls []string
}

func (h *WebHost) Start() error {
	return nil
}

func (h *WebHost) Stop(ctx context.Context) error {
	return nil
}

func (h *WebHost) Addresses() []string {
	if len(h.urls) == 0 {
		return []string{"http://localhost:8080"}
	}
	return h.urls
}

var _ abstract.WebHost = (*WebHost)(nil)
