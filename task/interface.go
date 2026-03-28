package task

import (
	"context"
	"time"

	"github.com/linuxerlv/gonest/core/abstract"
)

// ============================================================
//                    CronScheduler Interface
// ============================================================

// CronScheduler 定时任务调度器接口
type CronScheduler interface {
	// AddJob 添加定时任务
	// cronExpr: cron 表达式 (如 "0 * * * *" 每小时执行)
	// name: 任务名称
	// handler: 任务处理函数
	AddJob(cronExpr string, name string, handler JobHandler) error

	// AddIntervalJob 添加间隔任务
	AddIntervalJob(interval time.Duration, name string, handler JobHandler) error

	// RemoveJob 移除任务
	RemoveJob(name string) error

	// RunJob 立即执行一次任务
	RunJob(name string) error

	// Start 启动调度器
	Start() error

	// Stop 停止调度器
	Stop(ctx context.Context) error

	// Jobs 获取所有任务
	Jobs() []JobInfo
}

// JobHandler 任务处理函数
type JobHandler func(ctx context.Context) error

// JobInfo 任务信息
type JobInfo struct {
	Name       string
	Schedule   string
	NextRun    time.Time
	LastRun    time.Time
	Running    bool
	RunCount   int64
	ErrorCount int64
}

// ============================================================
//                    TaskQueue Interface
// ============================================================

// TaskQueue 后台任务队列接口
type TaskQueue interface {
	// Enqueue 入队任务
	Enqueue(task *QueueTask, opts ...TaskOption) error

	// EnqueueDelayed 延迟入队
	EnqueueDelayed(task *QueueTask, delay time.Duration, opts ...TaskOption) error

	// EnqueueAt 指定时间入队
	EnqueueAt(task *QueueTask, at time.Time, opts ...TaskOption) error

	// RegisterHandler 注册任务处理器
	RegisterHandler(taskType string, handler TaskHandler) error

	// Start 启动消费者
	Start(ctx context.Context) error

	// Stop 停止消费者
	Stop(ctx context.Context) error

	// Stats 获取队列统计
	Stats() QueueStats

	// Name 队列名称
	Name() string
}

// QueueTask 队列任务
type QueueTask struct {
	ID        string            // 任务ID
	Type      string            // 任务类型
	Payload   []byte            // 任务负载
	Priority  int               // 优先级
	Queue     string            // 队列名称
	Metadata  map[string]string // 元数据
	CreatedAt time.Time         // 创建时间
}

// TaskHandler 任务处理函数
type TaskHandler func(ctx context.Context, task *QueueTask) error

// TaskOption 任务选项
type TaskOption func(*TaskOptions)

// TaskOptions 任务选项
type TaskOptions struct {
	MaxRetry  int           // 最大重试次数
	Timeout   time.Duration // 超时时间
	Unique    bool          // 是否唯一任务
	Retention time.Duration // 结果保留时间
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

// QueueStats 队列统计
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

// ============================================================
//                    QueueFactory Interface
// ============================================================

// QueueFactory 队列工厂接口
type QueueFactory interface {
	// CreateQueue 创建队列
	CreateQueue(name string, opts QueueConfig) (TaskQueue, error)

	// CreateScheduler 创建调度器
	CreateScheduler(opts SchedulerConfig) (CronScheduler, error)
}

// QueueConfig 队列配置
type QueueConfig struct {
	Workers     int            // 工作协程数
	MaxRetry    int            // 最大重试次数
	Concurrency int            // 并发数
	Timeout     time.Duration  // 超时时间
	Queues      map[string]int // 队列优先级
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Timezone    string        // 时区
	Concurrency int           // 并发数
	Timeout     time.Duration // 超时时间
}

// ============================================================
//                    Result Interface
// ============================================================

// TaskResult 任务结果
type TaskResult struct {
	TaskID string
	Status string
	Result []byte
	Error  string
	At     time.Time
}

// ResultStore 结果存储接口
type ResultStore interface {
	// Set 设置结果
	Set(ctx context.Context, taskID string, result *TaskResult, ttl time.Duration) error

	// Get 获取结果
	Get(ctx context.Context, taskID string) (*TaskResult, error)

	// Delete 删除结果
	Delete(ctx context.Context, taskID string) error
}

// ============================================================
//                    Abstract Interface Aliases
// ============================================================

// Re-export abstract interfaces for users who want the interface-based API

// CronSchedulerAbstract aliases the abstract interface from core/abstract
type CronSchedulerAbstract = abstract.CronSchedulerAbstract

// TaskQueueAbstract aliases the abstract interface from core/abstract
type TaskQueueAbstract = abstract.TaskQueueAbstract

// QueueFactoryAbstract aliases the abstract interface from core/abstract
type QueueFactoryAbstract = abstract.QueueFactoryAbstract

// ResultStoreAbstract aliases the abstract interface from core/abstract
type ResultStoreAbstract = abstract.ResultStoreAbstract

// JobHandlerAbstract aliases the abstract handler type from core/abstract
type JobHandlerAbstract = abstract.JobHandlerAbstract

// TaskHandlerAbstract aliases the abstract handler type from core/abstract
type TaskHandlerAbstract = abstract.TaskHandlerAbstract

// TaskOptionAbstract aliases the abstract option type from core/abstract
type TaskOptionAbstract = abstract.TaskOptionAbstract

// TaskOptionsAbstract aliases the abstract options struct from core/abstract
type TaskOptionsAbstract = abstract.TaskOptionsAbstract

// QueueConfigAbstract aliases the abstract config struct from core/abstract
type QueueConfigAbstract = abstract.QueueConfigAbstract

// SchedulerConfigAbstract aliases the abstract config struct from core/abstract
type SchedulerConfigAbstract = abstract.SchedulerConfigAbstract

// WithMaxRetryAbstract aliases the abstract option function from core/abstract
var WithMaxRetryAbstract = abstract.WithMaxRetryAbstract

// WithTimeoutAbstract aliases the abstract option function from core/abstract
var WithTimeoutAbstract = abstract.WithTimeoutAbstract

// WithUniqueAbstract aliases the abstract option function from core/abstract
var WithUniqueAbstract = abstract.WithUniqueAbstract
