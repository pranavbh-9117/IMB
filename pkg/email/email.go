// Package email provides email delivery functionality for the IMB platform.
package email

import "context"

// Message defines the structure of an outbound email.
type Message struct {
	To      string
	Subject string
	Body    string
}

// EmailService defines operations for synchronous and asynchronous email delivery.
type EmailService interface {
	// Send dispatches the email synchronously. Returns an error on failure.
	Send(ctx context.Context, msg Message) error

	// SendAsync dispatches the email in a background goroutine. Never blocks.
	// Retries and timeouts are handled internally.
	SendAsync(ctx context.Context, msg Message)
}
