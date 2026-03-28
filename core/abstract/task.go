package abstract

import "context"
import "time"

// JobHandler 任务处理函数类型
type JobHandler func(ctx context.Context) error

// JobInfo 任务信息接口
type JobInfo interface {
	Name() string
	Schedule() string
	NextRun() time.Time
	LastRun() time.Time
	Running() bool
	RunCount() int64
	ErrorCount() int64
}

// CronScheduler 定时任务调度器接口
type CronScheduler interface {
	AddJob(cronExpr string, name string, handler JobHandler) error
	AddIntervalJob(interval time.Duration, name string, handler JobHandler) error
	RemoveJob(name string) error
	RunJob(name string) error
	Start() error
	Stop(ctx context.Context) error
	Jobs() []JobInfo
}

// TaskHandler 后台任务处理函数类型
type TaskHandler func(ctx context.Context, task *QueueTask) error

// QueueTask 队列任务接口
type QueueTask interface {
	ID() string
	Type() string
	Payload() []byte
	Priority() int
	Queue() string
	Metadata() map[string]string
	CreatedAt() time.Time
}

// QueueStats 队列统计接口
type QueueStats interface {
	Name() string
	Pending() int64
	Active() int64
	Scheduled() int64
	Retry() int64
	Processed() int64
	Failed() int64
	ProcessedAt() time.Time
}

// TaskQueue 后台任务队列接口
type TaskQueue interface {
	Enqueue(task QueueTask, opts ...TaskOption) error
	EnqueueDelayed(task QueueTask, delay time.Duration, opts ...TaskOption) error
	EnqueueAt(task QueueTask, at time.Time, opts ...TaskOption) error
	RegisterHandler(taskType string, handler TaskHandler) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stats() QueueStats
	Name() string
}

// TaskOption 任务选项函数类型
type TaskOption func(*TaskOptions)

// TaskOptions 任务选项
type TaskOptions struct {
	MaxRetry  int
	Timeout   time.Duration
	Unique    bool
	Retention time.Duration
}

// WithMaxRetry 设置最大重试次数
func WithMaxRetry(n int) TaskOption {
	return func(o *TaskOptions) { o.MaxRetry = n }
}

// WithTimeout 设置超时时间
func WithTimeout(d time.Duration) TaskOption {
	return func(o *TaskOptions) { o.Timeout = d }
}

// WithUnique 设置唯一任务
func WithUnique() TaskOption {
	return func(o *TaskOptions) { o.Unique = true }
}

// QueueConfig 队列配置
type QueueConfig struct {
	Workers     int
	MaxRetry    int
	Concurrency int
	Timeout     time.Duration
	Queues      map[string]int
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Timezone    string
	Concurrency int
	Timeout     time.Duration
}

// QueueFactory 队列工厂接口
type QueueFactory interface {
	CreateQueue(name string, opts QueueConfig) (TaskQueue, error)
	CreateScheduler(opts SchedulerConfig) (CronScheduler, error)
}

// TaskResult 任务结果接口
type TaskResult interface {
	TaskID() string
	Status() string
	Result() []byte
	Error() string
	At() time.Time
}

// ResultStore 结果存储接口
type ResultStore interface {
	Set(ctx context.Context, taskID string, result TaskResult, ttl time.Duration) error
	Get(ctx context.Context, taskID string) (TaskResult, error)
	Delete(ctx context.Context, taskID string) error
}
