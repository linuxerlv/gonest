package core

import (
	"github.com/linuxerlv/gonest/core/abstract"
)

// HostBuilder 主机构建器，实现 abstract.HostBuilderAbstract 接口
type HostBuilder struct {
	contentRootPath string
	environment     string
	args            []string
}

// NewHostBuilder 创建新的 HostBuilder
func NewHostBuilder() *HostBuilder {
	return &HostBuilder{
		environment: "development",
	}
}

// UseContentRoot 设置内容根目录
func (h *HostBuilder) UseContentRoot(path string) abstract.HostBuilderAbstract {
	h.contentRootPath = path
	return h
}

// UseEnvironment 设置环境
func (h *HostBuilder) UseEnvironment(env string) abstract.HostBuilderAbstract {
	h.environment = env
	return h
}

// ContentRoot 获取内容根目录
func (h *HostBuilder) ContentRoot() string {
	return h.contentRootPath
}

// Environment 获取环境
func (h *HostBuilder) Environment() string {
	return h.environment
}

// Args 获取命令行参数
func (h *HostBuilder) Args() []string {
	return h.args
}

func (h *HostBuilder) SetArgs(args []string) {
	h.args = args
}

var _ abstract.HostBuilderAbstract = (*HostBuilder)(nil)
