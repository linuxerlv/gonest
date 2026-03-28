//go:build !nomemory

package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
)

type RobfigCronAdapter struct {
	cron    *cron.Cron
	jobs    map[string]*cronJobInfo
	mu      sync.RWMutex
	running bool
}

type cronJobInfo struct {
	name     string
	schedule string
	entryID  cron.EntryID
	handler  JobHandler
	nextRun  time.Time
	lastRun  time.Time
	runCount int64
	errCount int64
}

func NewRobfigCronScheduler(opts ...RobfigCronOption) CronScheduler {
	options := &robfigCronOptions{
		timezone: time.Local,
		seconds:  false,
		logger:   nil,
	}
	for _, opt := range opts {
		opt(options)
	}

	configOpts := []cron.Option{}
	if options.timezone != nil {
		configOpts = append(configOpts, cron.WithLocation(options.timezone))
	}
	if options.seconds {
		configOpts = append(configOpts, cron.WithSeconds())
	}
	if options.logger != nil {
		configOpts = append(configOpts, cron.WithLogger(options.logger))
	}

	return &RobfigCronAdapter{
		cron: cron.New(configOpts...),
		jobs: make(map[string]*cronJobInfo),
	}
}

type robfigCronOptions struct {
	timezone *time.Location
	seconds  bool
	logger   cron.Logger
}

type RobfigCronOption func(*robfigCronOptions)

func WithTimezone(tz *time.Location) RobfigCronOption {
	return func(o *robfigCronOptions) { o.timezone = tz }
}

func WithSeconds() RobfigCronOption {
	return func(o *robfigCronOptions) { o.seconds = true }
}

func WithCronLogger(logger cron.Logger) RobfigCronOption {
	return func(o *robfigCronOptions) { o.logger = logger }
}

func (s *RobfigCronAdapter) AddJob(cronExpr string, name string, handler JobHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	wrappedHandler := func() {
		s.runJob(name, handler)
	}

	entryID, err := s.cron.AddFunc(cronExpr, wrappedHandler)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	s.jobs[name] = &cronJobInfo{
		name:     name,
		schedule: cronExpr,
		entryID:  entryID,
		handler:  handler,
	}

	return nil
}

func (s *RobfigCronAdapter) AddIntervalJob(interval time.Duration, name string, handler JobHandler) error {
	cronExpr := fmt.Sprintf("@every %v", interval)
	return s.AddJob(cronExpr, name, handler)
}

func (s *RobfigCronAdapter) runJob(name string, handler JobHandler) {
	s.mu.Lock()
	job, exists := s.jobs[name]
	if !exists {
		s.mu.Unlock()
		return
	}
	job.runCount++
	job.lastRun = time.Now()
	s.mu.Unlock()

	if err := handler(context.Background()); err != nil {
		s.mu.Lock()
		if j, ok := s.jobs[name]; ok {
			j.errCount++
		}
		s.mu.Unlock()
	}
}

func (s *RobfigCronAdapter) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	s.cron.Remove(job.entryID)
	delete(s.jobs, name)
	return nil
}

func (s *RobfigCronAdapter) RunJob(name string) error {
	s.mu.RLock()
	job, exists := s.jobs[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	go s.runJob(name, job.handler)
	return nil
}

func (s *RobfigCronAdapter) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.cron.Start()
	s.running = true
	return nil
}

func (s *RobfigCronAdapter) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctxCancel := s.cron.Stop()
	<-ctxCancel.Done()

	s.running = false
	return nil
}

func (s *RobfigCronAdapter) Jobs() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		entry := s.cron.Entry(job.entryID)
		result = append(result, JobInfo{
			Name:       job.name,
			Schedule:   job.schedule,
			NextRun:    entry.Next,
			LastRun:    job.lastRun,
			Running:    false,
			RunCount:   job.runCount,
			ErrorCount: job.errCount,
		})
	}
	return result
}
