package handlers

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestVerifySendGridSignature(t *testing.T) {
	secret := "sg-secret"
	ts := "1714412000"
	body := `[{"event":"bounce"}]`
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ts + "." + body))
	sig := hex.EncodeToString(mac.Sum(nil))

	if !verifySendGridSignature(secret, ts, body, sig) {
		t.Fatal("expected valid sendgrid signature")
	}
	if verifySendGridSignature(secret, ts, body, "bad-signature") {
		t.Fatal("expected invalid sendgrid signature")
	}
}

func TestVerifyTwilioSignature(t *testing.T) {
	secret := "tw-secret"
	body := "MessageSid=SM123&MessageStatus=delivered"
	mac := hmac.New(sha1.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !verifyTwilioSignature(secret, body, sig) {
		t.Fatal("expected valid twilio signature")
	}
	if verifyTwilioSignature(secret, body, "bad-signature") {
		t.Fatal("expected invalid twilio signature")
	}
}

func TestCallbackFailureTaxonomy(t *testing.T) {
	if reason, suppress := sendGridFailureTaxonomy("bounce"); reason != "bounce" || !suppress {
		t.Fatalf("unexpected sendgrid taxonomy: %s %v", reason, suppress)
	}
	if reason, suppress := twilioFailureTaxonomy("undelivered"); reason != "undelivered" || !suppress {
		t.Fatalf("unexpected twilio taxonomy: %s %v", reason, suppress)
	}
}
