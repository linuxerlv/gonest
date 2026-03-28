package abstract

import "context"
import "time"

// JobHandlerAbstract 任务处理函数类型
type JobHandlerAbstract func(ctx context.Context) error

// JobInfoAbstract 任务信息接口
type JobInfoAbstract interface {
	Name() string
	Schedule() string
	NextRun() time.Time
	LastRun() time.Time
	Running() bool
	RunCount() int64
	ErrorCount() int64
}

// CronSchedulerAbstract 定时任务调度器接口
type CronSchedulerAbstract interface {
	AddJob(cronExpr string, name string, handler JobHandlerAbstract) error
	AddIntervalJob(interval time.Duration, name string, handler JobHandlerAbstract) error
	RemoveJob(name string) error
	RunJob(name string) error
	Start() error
	Stop(ctx context.Context) error
	Jobs() []JobInfoAbstract
}

// TaskHandlerAbstract 后台任务处理函数类型
type TaskHandlerAbstract func(ctx context.Context, task *QueueTaskAbstract) error

// QueueTaskAbstract 队列任务接口
type QueueTaskAbstract interface {
	ID() string
	Type() string
	Payload() []byte
	Priority() int
	Queue() string
	Metadata() map[string]string
	CreatedAt() time.Time
}

// QueueStatsAbstract 队列统计接口
type QueueStatsAbstract interface {
	Name() string
	Pending() int64
	Active() int64
	Scheduled() int64
	Retry() int64
	Processed() int64
	Failed() int64
	ProcessedAt() time.Time
}

// TaskQueueAbstract 后台任务队列接口
type TaskQueueAbstract interface {
	Enqueue(task QueueTaskAbstract, opts ...TaskOptionAbstract) error
	EnqueueDelayed(task QueueTaskAbstract, delay time.Duration, opts ...TaskOptionAbstract) error
	EnqueueAt(task QueueTaskAbstract, at time.Time, opts ...TaskOptionAbstract) error
	RegisterHandler(taskType string, handler TaskHandlerAbstract) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Stats() QueueStatsAbstract
	Name() string
}

// TaskOptionAbstract 任务选项函数类型
type TaskOptionAbstract func(*TaskOptionsAbstract)

// TaskOptionsAbstract 任务选项
type TaskOptionsAbstract struct {
	MaxRetry  int
	Timeout   time.Duration
	Unique    bool
	Retention time.Duration
}

// WithMaxRetryAbstract 设置最大重试次数
func WithMaxRetryAbstract(n int) TaskOptionAbstract {
	return func(o *TaskOptionsAbstract) { o.MaxRetry = n }
}

// WithTimeoutAbstract 设置超时时间
func WithTimeoutAbstract(d time.Duration) TaskOptionAbstract {
	return func(o *TaskOptionsAbstract) { o.Timeout = d }
}

// WithUniqueAbstract 设置唯一任务
func WithUniqueAbstract() TaskOptionAbstract {
	return func(o *TaskOptionsAbstract) { o.Unique = true }
}

// QueueConfigAbstract 队列配置
type QueueConfigAbstract struct {
	Workers     int
	MaxRetry    int
	Concurrency int
	Timeout     time.Duration
	Queues      map[string]int
}

// SchedulerConfigAbstract 调度器配置
type SchedulerConfigAbstract struct {
	Timezone    string
	Concurrency int
	Timeout     time.Duration
}

// QueueFactoryAbstract 队列工厂接口
type QueueFactoryAbstract interface {
	CreateQueue(name string, opts QueueConfigAbstract) (TaskQueueAbstract, error)
	CreateScheduler(opts SchedulerConfigAbstract) (CronSchedulerAbstract, error)
}

// TaskResultAbstract 任务结果接口
type TaskResultAbstract interface {
	TaskID() string
	Status() string
	Result() []byte
	Error() string
	At() time.Time
}

// ResultStoreAbstract 结果存储接口
type ResultStoreAbstract interface {
	Set(ctx context.Context, taskID string, result TaskResultAbstract, ttl time.Duration) error
	Get(ctx context.Context, taskID string) (TaskResultAbstract, error)
	Delete(ctx context.Context, taskID string) error
}
