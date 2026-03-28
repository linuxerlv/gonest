package core

import (
	"os"
	"sync"

	"github.com/linuxerlv/gonest/core/abstract"
)

// Env 环境变量实现
type Env struct {
	values map[string]string
	mu     sync.RWMutex
}

// NewEnv 创建环境变量实例，自动加载系统环境变量
func NewEnv() *Env {
	env := &Env{
		values: make(map[string]string),
	}
	// 加载系统环境变量
	for _, pair := range os.Environ() {
		// os.Environ() 返回 "key=value" 格式
		for i := 0; i < len(pair); i++ {
			if pair[i] == '=' {
				env.values[pair[:i]] = pair[i+1:]
				break
			}
		}
	}
	return env
}

// Get 获取环境变量
func (e *Env) Get(key string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.values[key]
}

// GetOrDefault 获取环境变量，不存在返回默认值
func (e *Env) GetOrDefault(key, defaultValue string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if v, ok := e.values[key]; ok {
		return v
	}
	return defaultValue
}

// Has 检查环境变量是否存在
func (e *Env) Has(key string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	_, ok := e.values[key]
	return ok
}

// All 返回所有环境变量
func (e *Env) All() map[string]string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make(map[string]string, len(e.values))
	for k, v := range e.values {
		result[k] = v
	}
	return result
}

// Set 设置环境变量
func (e *Env) Set(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.values[key] = value
	os.Setenv(key, value)
}

// Unset 删除环境变量
func (e *Env) Unset(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.values, key)
	os.Unsetenv(key)
}

var _ abstract.EnvAbstract = (*Env)(nil)
