package workerpool_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pranavbh-9117/IMB/internal/workerpool"
)

type testLogger struct{}

func (testLogger) Info(context.Context, string, ...any)  {}
func (testLogger) Error(context.Context, string, ...any) {}

func TestWorkerPool_SubmitAndExecute(t *testing.T) {
	p, err := workerpool.New(workerpool.Options{
		WorkersCount: 4,
		QueueSize:    10,
		Logger:       testLogger{},
	})
	if err != nil {
		t.Fatalf("unexpected error creating pool: %v", err)
	}

	var count atomic.Int32
	var wg sync.WaitGroup

	numJobs := 20
	wg.Add(numJobs)

	for i := 0; i < numJobs; i++ {
		err := p.Submit(context.Background(), workerpool.JobFunc(func(ctx context.Context) error {
			defer wg.Done()
			count.Add(1)
			time.Sleep(5 * time.Millisecond)
			return nil
		}))
		if err != nil {
			t.Errorf("unexpected submit error: %v", err)
		}
	}

	wg.Wait()

	if count.Load() != int32(numJobs) {
		t.Errorf("expected %d completed jobs, got %d", numJobs, count.Load())
	}

	metrics := p.Metrics()
	if metrics.CompletedJobs != int64(numJobs) {
		t.Errorf("expected metrics completed=%d, got %d", numJobs, metrics.CompletedJobs)
	}

	if err := p.Shutdown(context.Background()); err != nil {
		t.Errorf("unexpected shutdown error: %v", err)
	}
}

func TestWorkerPool_SubmitAfterShutdown(t *testing.T) {
	p, _ := workerpool.New(workerpool.Options{
		WorkersCount: 2,
		QueueSize:    5,
		Logger:       testLogger{},
	})

	_ = p.Shutdown(context.Background())

	err := p.Submit(context.Background(), workerpool.JobFunc(func(ctx context.Context) error { return nil }))
	if !errors.Is(err, workerpool.ErrPoolClosed) {
		t.Errorf("expected ErrPoolClosed, got %v", err)
	}
}

func TestWorkerPool_QueuedJobCancellation(t *testing.T) {
	p, _ := workerpool.New(workerpool.Options{
		WorkersCount: 1, // Only 1 worker to ensure queuing
		QueueSize:    10,
		Logger:       testLogger{},
	})

	blocker := make(chan struct{})
	_ = p.Submit(context.Background(), workerpool.JobFunc(func(ctx context.Context) error {
		<-blocker
		return nil
	}))

	canceledCtx, cancel := context.WithCancel(context.Background())

	var executed atomic.Bool
	_ = p.Submit(canceledCtx, workerpool.JobFunc(func(ctx context.Context) error {
		executed.Store(true)
		return nil
	}))
	cancel() // Cancel after enqueued into queue buffer

	close(blocker) // Let worker finish first job

	_ = p.Shutdown(context.Background())

	if executed.Load() {
		t.Errorf("queued job executed despite canceled context")
	}

	metrics := p.Metrics()
	if metrics.FailedJobs < 1 {
		t.Errorf("expected at least 1 failed job in metrics due to cancellation, got %d", metrics.FailedJobs)
	}
}

func TestWorkerPool_PanicHandling(t *testing.T) {
	p, _ := workerpool.New(workerpool.Options{
		WorkersCount: 2,
		QueueSize:    5,
		Logger:       testLogger{},
	})

	var wg sync.WaitGroup
	wg.Add(1)

	_ = p.Submit(context.Background(), workerpool.JobFunc(func(ctx context.Context) error {
		defer wg.Done()
		panic("intentional test panic")
	}))

	wg.Wait()

	// Pool should still be alive and accept new jobs
	wg.Add(1)
	err := p.Submit(context.Background(), workerpool.JobFunc(func(ctx context.Context) error {
		defer wg.Done()
		return nil
	}))
	if err != nil {
		t.Errorf("unexpected error submitting job after panic: %v", err)
	}

	wg.Wait()
	_ = p.Shutdown(context.Background())

	metrics := p.Metrics()
	if metrics.FailedJobs != 1 {
		t.Errorf("expected 1 failed job from panic, got %d", metrics.FailedJobs)
	}
}

func TestWorkerPool_GroupCoordination(t *testing.T) {
	p, _ := workerpool.New(workerpool.Options{
		WorkersCount: 3,
		QueueSize:    10,
		Logger:       testLogger{},
	})

	g := p.NewGroup(context.Background())

	expectedErr := errors.New("group test failure")

	_ = g.Submit(workerpool.JobFunc(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return expectedErr
	}))

	_ = g.Submit(workerpool.JobFunc(func(ctx context.Context) error {
		return nil
	}))

	err := g.Wait()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected group error %v, got %v", expectedErr, err)
	}

	_ = p.Shutdown(context.Background())
}
