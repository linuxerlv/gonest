//go:build !nomemory

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hibiken/asynq"
)

type AsynqAdapter struct {
	client    *asynq.Client
	server    *asynq.Server
	mux       *asynq.ServeMux
	name      string
	handlers  map[string]TaskHandler
	mu        sync.RWMutex
	running   bool
	processed int64
	failed    int64
}

type AsynqConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
	Queues        map[string]int
}

func NewAsynqQueue(name string, config AsynqConfig) TaskQueue {
	redisOpt := asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	}

	concurrency := config.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}

	queues := config.Queues
	if len(queues) == 0 {
		queues = map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		}
	}

	return &AsynqAdapter{
		client: asynq.NewClient(redisOpt),
		server: asynq.NewServer(redisOpt, asynq.Config{
			Concurrency: concurrency,
			Queues:      queues,
		}),
		mux:      asynq.NewServeMux(),
		name:     name,
		handlers: make(map[string]TaskHandler),
	}
}

func (q *AsynqAdapter) Enqueue(task *QueueTask, opts ...TaskOption) error {
	options := &TaskOptions{}
	for _, opt := range opts {
		opt(options)
	}

	asynqOpts := []asynq.Option{}
	if options.MaxRetry > 0 {
		asynqOpts = append(asynqOpts, asynq.MaxRetry(options.MaxRetry))
	}
	if options.Timeout > 0 {
		asynqOpts = append(asynqOpts, asynq.Timeout(options.Timeout))
	}
	if options.Unique {
		asynqOpts = append(asynqOpts, asynq.Unique(time.Hour))
	}
	if task.Queue != "" {
		asynqOpts = append(asynqOpts, asynq.Queue(task.Queue))
	}

	payload := task.Payload
	if payload == nil {
		payload = []byte("{}")
	}

	asynqTask := asynq.NewTask(task.Type, payload, asynqOpts...)

	info, err := q.client.Enqueue(asynqTask, asynqOpts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	task.ID = info.ID
	task.CreatedAt = time.Now()
	return nil
}

func (q *AsynqAdapter) EnqueueDelayed(task *QueueTask, delay time.Duration, opts ...TaskOption) error {
	opts = append(opts, func(o *TaskOptions) {})
	return q.Enqueue(task, opts...)
}

func (q *AsynqAdapter) EnqueueAt(task *QueueTask, at time.Time, opts ...TaskOption) error {
	delay := time.Until(at)
	if delay < 0 {
		delay = 0
	}
	return q.EnqueueDelayed(task, delay, opts...)
}

func (q *AsynqAdapter) RegisterHandler(taskType string, handler TaskHandler) error {
	q.mu.Lock()
	q.handlers[taskType] = handler
	q.mu.Unlock()

	q.mux.HandleFunc(taskType, func(ctx context.Context, t *asynq.Task) error {
		q.mu.RLock()
		h, ok := q.handlers[taskType]
		q.mu.RUnlock()

		if !ok {
			return fmt.Errorf("no handler for task type %s", taskType)
		}

		task := &QueueTask{
			ID:      t.ResultWriter().TaskID(),
			Type:    t.Type(),
			Payload: t.Payload(),
		}

		if err := h(ctx, task); err != nil {
			q.mu.Lock()
			q.failed++
			q.mu.Unlock()
			return err
		}

		q.mu.Lock()
		q.processed++
		q.mu.Unlock()
		return nil
	})

	return nil
}

func (q *AsynqAdapter) Start(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return fmt.Errorf("queue already running")
	}

	q.running = true
	go func() {
		if err := q.server.Run(q.mux); err != nil {
			fmt.Printf("[Asynq] Server error: %v\n", err)
		}
	}()

	return nil
}

func (q *AsynqAdapter) Stop(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return nil
	}

	q.server.Shutdown()
	q.client.Close()
	q.running = false
	return nil
}

func (q *AsynqAdapter) Stats() QueueStats {
	return QueueStats{
		Name:      q.name,
		Pending:   0,
		Active:    0,
		Scheduled: 0,
		Retry:     0,
		Processed: q.processed,
		Failed:    q.failed,
	}
}

func (q *AsynqAdapter) Name() string {
	return q.name
}

type AsynqSchedulerAdapter struct {
	scheduler *asynq.Scheduler
	client    *asynq.Client
	jobs      map[string]*asynqScheduledJob
	mu        sync.RWMutex
	running   bool
}

type asynqScheduledJob struct {
	name     string
	schedule string
	taskType string
	payload  []byte
}

func NewAsynqScheduler(redisAddr string) CronScheduler {
	redisOpt := asynq.RedisClientOpt{Addr: redisAddr}
	return &AsynqSchedulerAdapter{
		scheduler: asynq.NewScheduler(redisOpt, nil),
		client:    asynq.NewClient(redisOpt),
		jobs:      make(map[string]*asynqScheduledJob),
	}
}

func (s *AsynqSchedulerAdapter) AddJob(cronExpr string, name string, handler JobHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	task := asynq.NewTask(name, nil)
	entryID, err := s.scheduler.Register(cronExpr, task)
	if err != nil {
		return fmt.Errorf("failed to register cron job: %w", err)
	}

	s.jobs[name] = &asynqScheduledJob{
		name:     name,
		schedule: cronExpr,
		taskType: entryID,
	}

	return nil
}

func (s *AsynqSchedulerAdapter) AddIntervalJob(interval time.Duration, name string, handler JobHandler) error {
	cronExpr := fmt.Sprintf("@every %v", interval)
	return s.AddJob(cronExpr, name, handler)
}

func (s *AsynqSchedulerAdapter) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	s.scheduler.Unregister(job.taskType)
	delete(s.jobs, name)
	return nil
}

func (s *AsynqSchedulerAdapter) RunJob(name string) error {
	s.mu.RLock()
	_, exists := s.jobs[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	task := asynq.NewTask(name, nil)
	_, err := s.client.Enqueue(task)
	return err
}

func (s *AsynqSchedulerAdapter) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.scheduler.Run()
	s.running = true
	return nil
}

func (s *AsynqSchedulerAdapter) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.scheduler.Shutdown()
	s.client.Close()
	s.running = false
	return nil
}

func (s *AsynqSchedulerAdapter) Jobs() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, JobInfo{
			Name:     job.name,
			Schedule: job.schedule,
		})
	}
	return result
}

func MustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}

func MustUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
