package email

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wneessen/go-mail"

	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/logger"
	"github.com/pranavbh-9117/IMB/pkg/retry"
)

type mailSender struct {
	cfg      config.SMTPConfig
	retryCfg retry.Config
	sendFunc func(ctx context.Context, msg Message) error
}

// NewMailSender creates a new production-ready SMTP EmailService.
func NewMailSender(cfg config.SMTPConfig) EmailService {
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	initialDelay := cfg.InitialDelay
	if initialDelay <= 0 {
		initialDelay = 500 * time.Millisecond
	}
	maxDelay := cfg.MaxDelay
	if maxDelay <= 0 {
		maxDelay = 10 * time.Second
	}
	multiplier := cfg.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	s := &mailSender{
		cfg: cfg,
		retryCfg: retry.Config{
			MaxAttempts:  maxAttempts,
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			Multiplier:   multiplier,
			ShouldRetry:  retry.IsSMTPTransientError,
		},
	}
	s.sendFunc = s.sendViaGoMail
	return s
}

// Send dispatches an email synchronously with configured retry parameters.
func (s *mailSender) Send(ctx context.Context, msg Message) error {
	return retry.Do(ctx, func() error {
		return s.sendFunc(ctx, msg)
	}, s.retryCfg)
}

// SendAsync dispatches an email in a background goroutine detached from request cancellation.
func (s *mailSender) SendAsync(ctx context.Context, msg Message) {
	bgCtx := context.WithoutCancel(ctx)
	timeout := s.cfg.SendTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	go func() {
		timeoutCtx, cancel := context.WithTimeout(bgCtx, timeout)
		defer cancel()

		if err := s.Send(timeoutCtx, msg); err != nil {
			logger.Error(timeoutCtx, "email: async send failed after retries",
				"to", msg.To,
				"subject", msg.Subject,
				"error", err.Error(),
			)
		}
	}()
}

func (s *mailSender) sendViaGoMail(ctx context.Context, msg Message) error {
	port, err := strconv.Atoi(s.cfg.Port)
	if err != nil {
		return fmt.Errorf("invalid smtp port: %w", err)
	}

	c, err := mail.NewClient(s.cfg.Host,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(s.cfg.Username),
		mail.WithPassword(s.cfg.Password),
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	m := mail.NewMsg()
	if err := m.From(s.cfg.From); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := m.To(sanitizeHeader(msg.To)); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}
	m.Subject(sanitizeHeader(msg.Subject))
	m.SetBodyString(mail.TypeTextPlain, sanitizeBody(msg.Body))

	if err := c.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("smtp send failed: %w", err)
	}
	return nil
}

// sanitizeHeader strips carriage returns and newlines to prevent SMTP header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}

// sanitizeBody strips carriage returns to prevent CRLF injection.
func sanitizeBody(s string) string {
	return strings.ReplaceAll(s, "\r", "")
}
