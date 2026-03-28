package task

import (
	"context"
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

type CronScheduler interface {
	AddJob(cronExpr string, name string, handler JobHandler) error
	AddIntervalJob(interval time.Duration, name string, handler JobHandler) error
	RemoveJob(name string) error
	RunJob(name string) error
	Start() error
	Stop(ctx context.Context) error
	Jobs() []JobInfo
}

type JobHandler func(ctx context.Context) error

type JobInfo struct {
	Name       string
	Schedule   string
	NextRun    time.Time
	LastRun    time.Time
	Running    bool
	RunCount   int64
	ErrorCount int64
}

type TaskQueue interface {
	Enqueue(task *QueueTask, opts ...TaskOption) error
	EnqueueDelayed(task *QueueTask, delay time.Duration, opts ...TaskOption) error
	EnqueueAt(task *QueueTask, at time.Time, opts ...TaskOption) error
	RegisterHandler(taskType string, handler TaskHandler) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stats() QueueStats
	Name() string
}

type QueueTask struct {
	ID        string
	Type      string
	Payload   []byte
	Priority  int
	Queue     string
	Metadata  map[string]string
	CreatedAt time.Time
}

type TaskHandler func(ctx context.Context, task *QueueTask) error

type TaskOption func(*TaskOptions)

type TaskOptions struct {
	MaxRetry  int
	Timeout   time.Duration
	Unique    bool
	Retention time.Duration
}

func WithMaxRetry(n int) TaskOption {
	return func(o *TaskOptions) { o.MaxRetry = n }
}

func WithTimeout(d time.Duration) TaskOption {
	return func(o *TaskOptions) { o.Timeout = d }
}

func WithUnique() TaskOption {
	return func(o *TaskOptions) { o.Unique = true }
}

type QueueStats struct {
	Name        string
	Pending     int64
	Active      int64
	Scheduled   int64
	Retry       int64
	Processed   int64
	Failed      int64
	ProcessedAt time.Time
}

type QueueFactory interface {
	CreateQueue(name string, opts QueueConfig) (TaskQueue, error)
	CreateScheduler(opts SchedulerConfig) (CronScheduler, error)
}

type QueueConfig struct {
	Workers     int
	MaxRetry    int
	Concurrency int
	Timeout     time.Duration
	Queues      map[string]int
}

type SchedulerConfig struct {
	Timezone    string
	Concurrency int
	Timeout     time.Duration
}

type TaskResult struct {
	TaskID string
	Status string
	Result []byte
	Error  string
	At     time.Time
}

type ResultStore interface {
	Set(ctx context.Context, taskID string, result *TaskResult, ttl time.Duration) error
	Get(ctx context.Context, taskID string) (*TaskResult, error)
	Delete(ctx context.Context, taskID string) error
}

type CronSchedulerAlias = abstract.CronScheduler
type TaskQueueAlias = abstract.TaskQueue
type QueueFactoryAlias = abstract.QueueFactory
type ResultStoreAlias = abstract.ResultStore
type JobHandlerAlias = abstract.JobHandler
type TaskHandlerAlias = abstract.TaskHandler
type TaskOptionAlias = abstract.TaskOption
type TaskOptionsAlias = abstract.TaskOptions
type QueueConfigAlias = abstract.QueueConfig
type SchedulerConfigAlias = abstract.SchedulerConfig

var WithMaxRetryAlias = abstract.WithMaxRetry
var WithTimeoutAlias = abstract.WithTimeout
var WithUniqueAlias = abstract.WithUnique
