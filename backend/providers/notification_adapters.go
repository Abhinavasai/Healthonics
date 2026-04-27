package providers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DeliveryRequest struct {
	To      string
	Subject string
	Body    string
}

type EmailProvider interface {
	Send(ctx context.Context, req DeliveryRequest) error
}

type SMSProvider interface {
	Send(ctx context.Context, req DeliveryRequest) error
}

type SendGridAdapter struct {
	APIKey  string
	From    string
	BaseURL string
	Client  *http.Client
}

func NewSendGridAdapter(apiKey, from string) *SendGridAdapter {
	return &SendGridAdapter{
		APIKey:  strings.TrimSpace(apiKey),
		From:    strings.TrimSpace(from),
		BaseURL: "https://api.sendgrid.com",
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SendGridAdapter) Send(ctx context.Context, req DeliveryRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return errors.New("sendgrid: recipient is required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return errors.New("sendgrid: body is required")
	}
	if s.APIKey == "" {
		return errors.New("sendgrid: api key is required")
	}
	if s.From == "" {
		return errors.New("sendgrid: from address is required")
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	baseURL := strings.TrimRight(s.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.sendgrid.com"
	}

	payload := map[string]any{
		"personalizations": []map[string]any{
			{
				"to": []map[string]string{
					{"email": strings.TrimSpace(req.To)},
				},
				"subject": strings.TrimSpace(req.Subject),
			},
		},
		"from": map[string]string{"email": s.From},
		"content": []map[string]string{
			{"type": "text/plain", "value": req.Body},
		},
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v3/mail/send", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	res, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("sendgrid: status %d: %s", res.StatusCode, string(b))
	}
	return nil
}

type TwilioAdapter struct {
	AccountSID string
	AuthToken  string
	From       string
	BaseURL    string
	Client     *http.Client
}

func NewTwilioAdapter(accountSID, authToken, from string) *TwilioAdapter {
	return &TwilioAdapter{
		AccountSID: strings.TrimSpace(accountSID),
		AuthToken:  strings.TrimSpace(authToken),
		From:       strings.TrimSpace(from),
		BaseURL:    "https://api.twilio.com",
		Client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TwilioAdapter) Send(ctx context.Context, req DeliveryRequest) error {
	if strings.TrimSpace(req.To) == "" {
		return errors.New("twilio: recipient is required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return errors.New("twilio: body is required")
	}
	if t.AccountSID == "" || t.AuthToken == "" {
		return errors.New("twilio: account credentials are required")
	}
	if t.From == "" {
		return errors.New("twilio: from number is required")
	}
	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	baseURL := strings.TrimRight(t.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.twilio.com"
	}

	form := url.Values{}
	form.Set("To", req.To)
	form.Set("From", t.From)
	form.Set("Body", req.Body)

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", baseURL, url.PathEscape(t.AccountSID))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	encoded := base64.StdEncoding.EncodeToString([]byte(t.AccountSID + ":" + t.AuthToken))
	httpReq.Header.Set("Authorization", "Basic "+encoded)

	res, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("twilio: status %d: %s", res.StatusCode, string(b))
	}
	return nil
}

