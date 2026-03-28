package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/linuxerlv/gonest/config"
	"github.com/linuxerlv/gonest/core/abstract"
	"github.com/linuxerlv/gonest/logger"
)

type HostApplication struct {
	config   config.Config
	env      abstract.Env
	services *ServiceCollection
	logger   logger.Logger
	values   map[string]any
	mu       sync.RWMutex
	started  bool
	stopped  bool
	stopCh   chan struct{}
	runCh    chan error
}

func NewHostApplication() *HostApplication {
	return &HostApplication{
		env:      NewEnv(),
		services: NewServiceCollection(),
		values:   make(map[string]any),
		stopCh:   make(chan struct{}),
		runCh:    make(chan error, 1),
	}
}

func (h *HostApplication) Services() abstract.ServiceCollection {
	return h.services
}

func (h *HostApplication) Configuration() abstract.Config {
	if h.config != nil {
		return NewConfigAdapter(h.config)
	}
	return nil
}

func (h *HostApplication) Environment() abstract.Env {
	return h.env
}

func (h *HostApplication) Logging() abstract.Logger {
	if h.logger != nil {
		return NewLoggerAdapter(h.logger)
	}
	return NewLoggerAdapter(logger.GetGlobalLogger())
}

func (h *HostApplication) Run() error {
	if err := h.Start(); err != nil {
		return err
	}
	return h.WaitForShutdown()
}

func (h *HostApplication) RunAsync() <-chan error {
	go func() {
		h.runCh <- h.Run()
	}()
	return h.runCh
}

func (h *HostApplication) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return fmt.Errorf("application already started")
	}

	h.started = true
	return nil
}

func (h *HostApplication) StartAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- h.Start()
	}()
	return ch
}

func (h *HostApplication) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.started || h.stopped {
		return nil
	}

	h.stopped = true
	close(h.stopCh)

	return nil
}

func (h *HostApplication) Shutdown(ctx context.Context) error {
	return h.Stop()
}

func (h *HostApplication) WaitForShutdown() error {
	<-h.stopCh
	return nil
}

func (h *HostApplication) IsStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.started
}

func (h *HostApplication) IsStopped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stopped
}

var _ abstract.Application = (*HostApplication)(nil)
