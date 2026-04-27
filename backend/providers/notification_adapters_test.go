package providers

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendGridAdapter_Send_Success(t *testing.T) {
	var gotAuth string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/mail/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	adapter := NewSendGridAdapter("sg-key", "noreply@example.com")
	adapter.BaseURL = srv.URL
	if err := adapter.Send(context.Background(), DeliveryRequest{
		To:      "user@example.com",
		Subject: "Hello",
		Body:    "Test body",
	}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if gotAuth != "Bearer sg-key" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if !strings.Contains(gotBody, `"user@example.com"`) || !strings.Contains(gotBody, `"Test body"`) {
		t.Fatalf("unexpected request body: %s", gotBody)
	}
}

func TestSendGridAdapter_Send_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer srv.Close()

	adapter := NewSendGridAdapter("sg-key", "noreply@example.com")
	adapter.BaseURL = srv.URL
	err := adapter.Send(context.Background(), DeliveryRequest{
		To:      "user@example.com",
		Subject: "Hello",
		Body:    "Test body",
	})
	if err == nil || !strings.Contains(err.Error(), "status 400") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestTwilioAdapter_Send_Success(t *testing.T) {
	var gotAuth string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/Messages.json") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	adapter := NewTwilioAdapter("AC123", "tok123", "+15551112222")
	adapter.BaseURL = srv.URL
	if err := adapter.Send(context.Background(), DeliveryRequest{
		To:   "+15550001111",
		Body: "Reminder",
	}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("AC123:tok123"))
	if gotAuth != wantAuth {
		t.Fatalf("unexpected auth header: got %q want %q", gotAuth, wantAuth)
	}
	if !strings.Contains(gotBody, "To=%2B15550001111") || !strings.Contains(gotBody, "Body=Reminder") {
		t.Fatalf("unexpected twilio body: %s", gotBody)
	}
}

func TestTwilioAdapter_Send_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	adapter := NewTwilioAdapter("AC123", "tok123", "+15551112222")
	adapter.BaseURL = srv.URL
	err := adapter.Send(context.Background(), DeliveryRequest{
		To:   "+15550001111",
		Body: "Reminder",
	})
	if err == nil || !strings.Contains(err.Error(), "status 401") {
		t.Fatalf("expected status error, got %v", err)
	}
}

func TestAdapters_ValidateRequiredFields(t *testing.T) {
	sg := NewSendGridAdapter("", "")
	if err := sg.Send(context.Background(), DeliveryRequest{To: "x", Body: "y"}); err == nil {
		t.Fatal("expected sendgrid validation error")
	}

	tw := NewTwilioAdapter("", "", "")
	if err := tw.Send(context.Background(), DeliveryRequest{To: "x", Body: "y"}); err == nil {
		t.Fatal("expected twilio validation error")
	}
}

