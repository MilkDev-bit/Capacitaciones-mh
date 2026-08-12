package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := New(Config{
		APIKey:  "re_test_key",
		From:    "Capacitaciones <no-reply@example.com>",
		AppName: "Capacitaciones MH",
		AppURL:  "https://app.example.com",
	})
	c.endpoint = srv.URL
	return c
}

func TestSendBuildsResendPayload(t *testing.T) {
	var got map[string]any

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emails" {
			t.Errorf("path = %q, se esperaba /emails", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer re_test_key" {
			t.Errorf("Authorization = %q", auth)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	})

	err := c.Send(context.Background(), Message{
		To:      []string{"user@example.com"},
		Subject: "Hola",
		HTML:    "<p>hola</p>",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if got["from"] != "Capacitaciones <no-reply@example.com>" {
		t.Errorf("from = %v", got["from"])
	}
	if got["subject"] != "Hola" {
		t.Errorf("subject = %v", got["subject"])
	}
}

func TestSendPropagatesAPIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"dominio no verificado"}`))
	})

	err := c.Send(context.Background(), Message{To: []string{"a@b.com"}, Subject: "s", HTML: "h"})
	if err == nil {
		t.Fatal("se esperaba error de la API")
	}
	if !strings.Contains(err.Error(), "dominio no verificado") {
		t.Errorf("error = %v, debería incluir el detalle de Resend", err)
	}
}

func TestSendWithoutAPIKeyIsNotConfigured(t *testing.T) {
	c := New(Config{AppName: "X"})
	if c.Enabled() {
		t.Fatal("Enabled() = true sin API key")
	}
	if err := c.Send(context.Background(), Message{To: []string{"a@b.com"}}); err != ErrNotConfigured {
		t.Errorf("err = %v, se esperaba ErrNotConfigured", err)
	}
}

func TestSendBatchChunksAtHundred(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/emails/batch" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var batch []map[string]any
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &batch)
		if len(batch) > maxBatchSize {
			t.Errorf("lote de %d, máximo %d", len(batch), maxBatchSize)
		}
		w.WriteHeader(http.StatusOK)
	})

	msgs := make([]Message, 150)
	for i := range msgs {
		msgs[i] = Message{To: []string{"a@b.com"}, Subject: "s", HTML: "h"}
	}
	if err := c.SendBatch(context.Background(), msgs); err != nil {
		t.Fatalf("SendBatch() error = %v", err)
	}
	if calls != 2 {
		t.Errorf("llamadas = %d, se esperaban 2", calls)
	}
}

func TestTemplatesEscapeUserInput(t *testing.T) {
	c := New(Config{AppName: "MH", AppURL: "https://app.example.com"})

	msg := c.VerificationCode(`<script>alert(1)</script>`, "123456", 15)
	if strings.Contains(msg.HTML, "<script>") {
		t.Error("el nombre no fue escapado en VerificationCode")
	}
	if !strings.Contains(msg.HTML, "123456") {
		t.Error("el código no aparece en el cuerpo")
	}

	acc := c.ParticipantAccess("Ana", `Acme <b>`, `Curso "X"`, "TCK-1", "https://app.example.com/join")
	if strings.Contains(acc.HTML, "<b>") {
		t.Error("el nombre del comprador no fue escapado")
	}
}

func TestFormatMoney(t *testing.T) {
	cases := map[float64]string{
		0:          "0.00",
		999.5:      "999.50",
		1234.5:     "1,234.50",
		1234567.89: "1,234,567.89",
	}
	for in, want := range cases {
		if got := formatMoney(in); got != want {
			t.Errorf("formatMoney(%v) = %q, se esperaba %q", in, got, want)
		}
	}
}

func TestSafeURLRejectsNonHTTP(t *testing.T) {
	if got := safeURL("javascript:alert(1)"); got != "" {
		t.Errorf("safeURL() = %q, se esperaba cadena vacía", got)
	}
	if got := safeURL("https://ok.example.com/x"); got == "" {
		t.Error("safeURL() rechazó una URL https válida")
	}
}
