package task

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type memoryCronScheduler struct {
	jobs    map[string]*memoryJob
	mu      sync.RWMutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type memoryJob struct {
	name     string
	schedule string
	handler  JobHandler
	nextRun  time.Time
	lastRun  time.Time
	running  bool
	runCount int64
	errCount int64
	interval time.Duration
	cronExpr string
	ticker   *time.Ticker
	stop     chan struct{}
}

func NewMemoryCronScheduler() CronScheduler {
	return &memoryCronScheduler{
		jobs: make(map[string]*memoryJob),
	}
}

func (s *memoryCronScheduler) AddJob(cronExpr string, name string, handler JobHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	job := &memoryJob{
		name:     name,
		cronExpr: cronExpr,
		schedule: cronExpr,
		handler:  handler,
		stop:     make(chan struct{}),
	}

	s.jobs[name] = job

	if s.running {
		s.startJob(job)
	}

	return nil
}

func (s *memoryCronScheduler) AddIntervalJob(interval time.Duration, name string, handler JobHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	job := &memoryJob{
		name:     name,
		interval: interval,
		schedule: fmt.Sprintf("every %v", interval),
		handler:  handler,
		nextRun:  time.Now().Add(interval),
		stop:     make(chan struct{}),
	}

	s.jobs[name] = job

	if s.running {
		s.startJob(job)
	}

	return nil
}

func (s *memoryCronScheduler) startJob(job *memoryJob) {
	var ticker *time.Ticker
	if job.interval > 0 {
		ticker = time.NewTicker(job.interval)
	} else {
		ticker = time.NewTicker(parseCronInterval(job.cronExpr))
	}
	job.ticker = ticker

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-job.stop:
				return
			case <-ticker.C:
				s.runJob(job)
			}
		}
	}()
}

func (s *memoryCronScheduler) runJob(job *memoryJob) {
	job.running = true
	job.lastRun = time.Now()
	job.runCount++

	if err := job.handler(s.ctx); err != nil {
		job.errCount++
		log.Printf("[Cron] Job %s error: %v", job.name, err)
	}

	job.running = false
	if job.interval > 0 {
		job.nextRun = time.Now().Add(job.interval)
	}
}

func (s *memoryCronScheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	if job.ticker != nil {
		job.ticker.Stop()
	}
	close(job.stop)
	delete(s.jobs, name)

	return nil
}

func (s *memoryCronScheduler) RunJob(name string) error {
	s.mu.RLock()
	job, exists := s.jobs[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	go s.runJob(job)
	return nil
}

func (s *memoryCronScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.running = true

	for _, job := range s.jobs {
		s.startJob(job)
	}

	return nil
}

func (s *memoryCronScheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.cancel()

	for _, job := range s.jobs {
		if job.ticker != nil {
			job.ticker.Stop()
		}
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.running = false
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout stopping scheduler")
	}
}

func (s *memoryCronScheduler) Jobs() []JobInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]JobInfo, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, JobInfo{
			Name:       job.name,
			Schedule:   job.schedule,
			NextRun:    job.nextRun,
			LastRun:    job.lastRun,
			Running:    job.running,
			RunCount:   job.runCount,
			ErrorCount: job.errCount,
		})
	}
	return result
}

func parseCronInterval(expr string) time.Duration {
	return time.Minute
}

type memoryTaskQueue struct {
	name      string
	handlers  map[string]TaskHandler
	tasks     chan *queueTaskInternal
	workers   int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
	mu        sync.RWMutex
	processed int64
	failed    int64
	maxRetry  int
}

type queueTaskInternal struct {
	task     *QueueTask
	opts     TaskOptions
	attempts int
}

func NewMemoryTaskQueue(name string, workers int, bufferSize int) TaskQueue {
	return &memoryTaskQueue{
		name:     name,
		handlers: make(map[string]TaskHandler),
		tasks:    make(chan *queueTaskInternal, bufferSize),
		workers:  workers,
		maxRetry: 3,
	}
}

func (q *memoryTaskQueue) Enqueue(task *QueueTask, opts ...TaskOption) error {
	options := &TaskOptions{}
	for _, opt := range opts {
		opt(options)
	}

	task.CreatedAt = time.Now()

	select {
	case q.tasks <- &queueTaskInternal{task: task, opts: *options}:
		return nil
	default:
		return fmt.Errorf("queue %s is full", q.name)
	}
}

func (q *memoryTaskQueue) EnqueueDelayed(task *QueueTask, delay time.Duration, opts ...TaskOption) error {
	go func() {
		time.Sleep(delay)
		q.Enqueue(task, opts...)
	}()
	return nil
}

func (q *memoryTaskQueue) EnqueueAt(task *QueueTask, at time.Time, opts ...TaskOption) error {
	delay := time.Until(at)
	if delay < 0 {
		delay = 0
	}
	return q.EnqueueDelayed(task, delay, opts...)
}

func (q *memoryTaskQueue) RegisterHandler(taskType string, handler TaskHandler) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = handler
	return nil
}

func (q *memoryTaskQueue) Start(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.running {
		return fmt.Errorf("queue already running")
	}

	q.ctx, q.cancel = context.WithCancel(ctx)
	q.running = true

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}

	log.Printf("[TaskQueue] %s started with %d workers", q.name, q.workers)
	return nil
}

func (q *memoryTaskQueue) worker(id int) {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case item := <-q.tasks:
			q.processTask(item, id)
		}
	}
}

func (q *memoryTaskQueue) processTask(item *queueTaskInternal, workerID int) {
	q.mu.RLock()
	handler, ok := q.handlers[item.task.Type]
	q.mu.RUnlock()

	if !ok {
		log.Printf("[TaskQueue] %s no handler for type %s", q.name, item.task.Type)
		q.failed++
		return
	}

	var err error
	maxRetry := item.opts.MaxRetry
	if maxRetry == 0 {
		maxRetry = q.maxRetry
	}

	for i := 0; i <= maxRetry; i++ {
		if err = handler(q.ctx, item.task); err == nil {
			q.processed++
			return
		}
		log.Printf("[TaskQueue] %s worker-%d retry %d for %s: %v",
			q.name, workerID, i, item.task.ID, err)
		time.Sleep(time.Second * time.Duration(i+1))
	}

	q.failed++
	log.Printf("[TaskQueue] %s task %s failed after %d retries",
		q.name, item.task.ID, maxRetry)
}

func (q *memoryTaskQueue) Stop(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.running {
		return nil
	}

	q.cancel()

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		q.running = false
		log.Printf("[TaskQueue] %s stopped", q.name)
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout stopping queue %s", q.name)
	}
}

func (q *memoryTaskQueue) Stats() QueueStats {
	return QueueStats{
		Name:      q.name,
		Pending:   int64(len(q.tasks)),
		Processed: q.processed,
		Failed:    q.failed,
	}
}

func (q *memoryTaskQueue) Name() string {
	return q.name
}

type memoryQueueFactory struct{}

func NewMemoryQueueFactory() QueueFactory {
	return &memoryQueueFactory{}
}

func (f *memoryQueueFactory) CreateQueue(name string, opts QueueConfig) (TaskQueue, error) {
	workers := opts.Workers
	if workers <= 0 {
		workers = 5
	}
	return NewMemoryTaskQueue(name, workers, 1000), nil
}

func (f *memoryQueueFactory) CreateScheduler(opts SchedulerConfig) (CronScheduler, error) {
	return NewMemoryCronScheduler(), nil
}
