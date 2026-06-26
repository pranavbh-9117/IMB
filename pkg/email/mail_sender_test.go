package email

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pranavbh-9117/IMB/pkg/config"
)

func TestSend_Success(t *testing.T) {
	cfg := config.SMTPConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
	}

	sender := NewMailSender(cfg).(*mailSender)

	var attempts int32
	sender.sendFunc = func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}

	err := sender.Send(context.Background(), Message{To: "test@example.com"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestSend_RetryOnTransientError(t *testing.T) {
	cfg := config.SMTPConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
	}

	sender := NewMailSender(cfg).(*mailSender)

	var attempts int32
	sender.sendFunc = func(ctx context.Context, msg Message) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("450 greylisted try again later")
		}
		return nil
	}

	err := sender.Send(context.Background(), Message{To: "test@example.com"})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestSend_NoRetryOnPermanentError(t *testing.T) {
	cfg := config.SMTPConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
	}

	sender := NewMailSender(cfg).(*mailSender)

	var attempts int32
	permErr := errors.New("550 user unknown")
	sender.sendFunc = func(ctx context.Context, msg Message) error {
		atomic.AddInt32(&attempts, 1)
		return permErr
	}

	err := sender.Send(context.Background(), Message{To: "test@example.com"})
	if !errors.Is(err, permErr) {
		t.Fatalf("expected permanent error %v, got %v", permErr, err)
	}
	if atomic.LoadInt32(&attempts) != 1 {
		t.Fatalf("expected exactly 1 attempt on permanent error, got %d", attempts)
	}
}

func TestSendAsync_ContextDetachment(t *testing.T) {
	cfg := config.SMTPConfig{
		SendTimeout:  200 * time.Millisecond,
		MaxAttempts:  1,
		InitialDelay: 10 * time.Millisecond,
	}

	sender := NewMailSender(cfg).(*mailSender)

	done := make(chan struct{})
	sender.sendFunc = func(ctx context.Context, msg Message) error {
		select {
		case <-ctx.Done():
			t.Errorf("async context was unexpectedly cancelled")
		default:
		}
		close(done)
		return nil
	}

	reqCtx, cancel := context.WithCancel(context.Background())
	sender.SendAsync(reqCtx, Message{To: "async@example.com"})
	cancel()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for async send")
	}
}
