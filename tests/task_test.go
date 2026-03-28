package gonest

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/linuxerlv/gonest/task"
)

// ============================================================
//                 Memory CronScheduler Tests
// ============================================================

func TestMemoryCronScheduler_AddJob(t *testing.T) {
	scheduler := task.NewMemoryCronScheduler()

	var counter int32

	err := scheduler.AddIntervalJob(10*time.Millisecond, "test-job", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to add job: %v", err)
	}

	scheduler.Start()

	time.Sleep(55 * time.Millisecond)

	if atomic.LoadInt32(&counter) < 4 {
		t.Errorf("expected at least 4 executions, got %d", counter)
	}

	scheduler.Stop(context.Background())
}

func TestMemoryCronScheduler_RemoveJob(t *testing.T) {
	scheduler := task.NewMemoryCronScheduler()

	var counter int32

	scheduler.AddIntervalJob(10*time.Millisecond, "removable-job", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	scheduler.Start()
	time.Sleep(30 * time.Millisecond)

	scheduler.RemoveJob("removable-job")
	time.Sleep(30 * time.Millisecond)

	countAfterRemove := atomic.LoadInt32(&counter)
	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&counter) > countAfterRemove+1 {
		t.Error("job should have been removed")
	}

	scheduler.Stop(context.Background())
}

func TestMemoryCronScheduler_RunJob(t *testing.T) {
	scheduler := task.NewMemoryCronScheduler()

	var counter int32

	scheduler.AddIntervalJob(time.Hour, "manual-job", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	scheduler.RunJob("manual-job")
	time.Sleep(10 * time.Millisecond)

	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("expected 1 execution, got %d", counter)
	}

	scheduler.Stop(context.Background())
}

func TestMemoryCronScheduler_Jobs(t *testing.T) {
	scheduler := task.NewMemoryCronScheduler()

	scheduler.AddIntervalJob(time.Minute, "job1", func(ctx context.Context) error { return nil })
	scheduler.AddIntervalJob(time.Minute, "job2", func(ctx context.Context) error { return nil })

	jobs := scheduler.Jobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}

	scheduler.Stop(context.Background())
}

// ============================================================
//                 Memory TaskQueue Tests
// ============================================================

func TestMemoryTaskQueue_Enqueue(t *testing.T) {
	queue := task.NewMemoryTaskQueue("test", 2, 10)

	var processed int32

	queue.RegisterHandler("test-task", func(ctx context.Context, task *task.QueueTask) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	queue.Start(context.Background())

	for i := 0; i < 5; i++ {
		queue.Enqueue(&task.QueueTask{
			ID:   string(rune('A' + i)),
			Type: "test-task",
		})
	}

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&processed) != 5 {
		t.Errorf("expected 5 processed, got %d", processed)
	}

	queue.Stop(context.Background())
}

func TestMemoryTaskQueue_EnqueueDelayed(t *testing.T) {
	queue := task.NewMemoryTaskQueue("delayed-test", 1, 10)

	var processed int32

	queue.RegisterHandler("delayed-task", func(ctx context.Context, task *task.QueueTask) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	queue.Start(context.Background())

	queue.EnqueueDelayed(&task.QueueTask{
		ID:   "delayed-1",
		Type: "delayed-task",
	}, 50*time.Millisecond)

	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&processed) != 0 {
		t.Error("task executed too early")
	}

	time.Sleep(80 * time.Millisecond)

	if atomic.LoadInt32(&processed) != 1 {
		t.Errorf("expected 1 processed, got %d", processed)
	}

	queue.Stop(context.Background())
}

func TestMemoryTaskQueue_Stats(t *testing.T) {
	queue := task.NewMemoryTaskQueue("stats-test", 1, 10)

	queue.RegisterHandler("stats-task", func(ctx context.Context, task *task.QueueTask) error {
		return nil
	})

	queue.Start(context.Background())

	for i := 0; i < 3; i++ {
		queue.Enqueue(&task.QueueTask{
			ID:   string(rune(i)),
			Type: "stats-task",
		})
	}

	time.Sleep(50 * time.Millisecond)

	stats := queue.Stats()

	if stats.Name != "stats-test" {
		t.Errorf("expected name 'stats-test', got '%s'", stats.Name)
	}

	if stats.Processed != 3 {
		t.Errorf("expected 3 processed, got %d", stats.Processed)
	}

	queue.Stop(context.Background())
}

func TestMemoryTaskQueue_TaskOptions(t *testing.T) {
	queue := task.NewMemoryTaskQueue("options-test", 1, 10)

	var processed int32

	queue.RegisterHandler("options-task", func(ctx context.Context, task *task.QueueTask) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	queue.Start(context.Background())

	queue.Enqueue(&task.QueueTask{
		ID:   "opt-1",
		Type: "options-task",
	}, task.WithMaxRetry(2), task.WithTimeout(5*time.Second))

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&processed) != 1 {
		t.Errorf("expected 1 processed, got %d", processed)
	}

	queue.Stop(context.Background())
}

// ============================================================
//                 Memory QueueFactory Tests
// ============================================================

func TestMemoryQueueFactory_CreateQueue(t *testing.T) {
	factory := task.NewMemoryQueueFactory()

	queue, err := factory.CreateQueue("test-queue", task.QueueConfig{
		Workers: 5,
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	if queue.Name() != "test-queue" {
		t.Errorf("expected name 'test-queue', got '%s'", queue.Name())
	}
}

func TestMemoryQueueFactory_CreateScheduler(t *testing.T) {
	factory := task.NewMemoryQueueFactory()

	scheduler, err := factory.CreateScheduler(task.SchedulerConfig{})
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	if scheduler == nil {
		t.Error("expected scheduler to be non-nil")
	}
}

// ============================================================
//                 RobfigCronAdapter Tests
// ============================================================

func TestRobfigCronScheduler_AddJob(t *testing.T) {
	scheduler := task.NewRobfigCronScheduler()

	var counter int32

	err := scheduler.AddJob("*/1 * * * *", "test-job", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to add job: %v", err)
	}

	scheduler.Start()
	time.Sleep(100 * time.Millisecond)

	scheduler.Stop(context.Background())
}

func TestRobfigCronScheduler_CronExpression(t *testing.T) {
	scheduler := task.NewRobfigCronScheduler(task.WithSeconds())

	var counter int32

	err := scheduler.AddJob("*/1 * * * * *", "every-second", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("failed to add job: %v", err)
	}

	scheduler.Start()
	time.Sleep(2500 * time.Millisecond)

	if atomic.LoadInt32(&counter) < 2 {
		t.Errorf("expected at least 2 executions, got %d", counter)
	}

	scheduler.Stop(context.Background())
}

func TestRobfigCronScheduler_RemoveJob(t *testing.T) {
	scheduler := task.NewRobfigCronScheduler()

	var counter int32

	scheduler.AddJob("@every 10ms", "removable", func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	})

	scheduler.Start()
	time.Sleep(30 * time.Millisecond)

	scheduler.RemoveJob("removable")
	countAfterRemove := atomic.LoadInt32(&counter)

	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&counter) > countAfterRemove+1 {
		t.Error("job should have been removed")
	}

	scheduler.Stop(context.Background())
}

func TestRobfigCronScheduler_Jobs(t *testing.T) {
	scheduler := task.NewRobfigCronScheduler()

	scheduler.AddJob("@every 1m", "job1", func(ctx context.Context) error { return nil })
	scheduler.AddJob("@every 2m", "job2", func(ctx context.Context) error { return nil })

	jobs := scheduler.Jobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}

	scheduler.Stop(context.Background())
}
