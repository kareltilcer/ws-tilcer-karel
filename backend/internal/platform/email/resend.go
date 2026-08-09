// Package email sends transactional mail via Resend (the contact form). It is a
// thin wrapper over Resend's REST API; failures are returned to the caller,
// which decides how to handle them — contact treats a send failure as non-fatal
// (the message is already stored).
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Sender sends an email.
type Sender interface {
	Send(ctx context.Context, msg Message) error
}

// Message is a single outbound email.
type Message struct {
	From    string
	To      string
	Subject string
	Text    string
	ReplyTo string
}

// ResendClient talks to the Resend API.
type ResendClient struct {
	apiKey string
	client *http.Client
	url    string
}

// NewResendClient returns a Resend-backed Sender.
func NewResendClient(apiKey string) *ResendClient {
	return &ResendClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
		url:    "https://api.resend.com/emails",
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	ReplyTo string   `json:"reply_to,omitempty"`
}

// Send delivers msg via Resend. Returns an error on any non-2xx or transport
// failure.
func (c *ResendClient) Send(ctx context.Context, msg Message) error {
	body, err := json.Marshal(resendPayload{
		From:    msg.From,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Text:    msg.Text,
		ReplyTo: msg.ReplyTo,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("resend: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("resend: status %d: %s", resp.StatusCode, string(snippet))
	}
	return nil
}
