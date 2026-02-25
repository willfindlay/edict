package transcribe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientTranscribe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("expected multipart/form-data, got %s", r.Header.Get("Content-Type"))
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("get file: %v", err)
		}
		defer file.Close()

		data, _ := io.ReadAll(file)
		if len(data) == 0 {
			t.Error("expected non-empty file data")
		}

		prompt := r.FormValue("prompt")
		if prompt != "test prompt" {
			t.Errorf("expected prompt 'test prompt', got %q", prompt)
		}

		resp := inferenceResponse{Text: " Hello, world. "}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	// Strip http:// prefix since NewClient adds it
	addr := strings.TrimPrefix(srv.URL, "http://")
	client := NewClient(addr)

	text, err := client.Transcribe([]byte("fake wav data"), "test prompt")
	if err != nil {
		t.Fatalf("transcribe failed: %v", err)
	}

	if text != "Hello, world." {
		t.Errorf("expected 'Hello, world.', got %q", text)
	}
}

func TestClientTranscribeNoPrompt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}

		prompt := r.FormValue("prompt")
		if prompt != "" {
			t.Errorf("expected empty prompt, got %q", prompt)
		}

		resp := inferenceResponse{Text: "some text"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	client := NewClient(addr)

	text, err := client.Transcribe([]byte("fake wav"), "")
	if err != nil {
		t.Fatalf("transcribe failed: %v", err)
	}
	if text != "some text" {
		t.Errorf("expected 'some text', got %q", text)
	}
}

func TestClientTranscribeServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	client := NewClient(addr)

	_, err := client.Transcribe([]byte("data"), "")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error mentioning 500, got %q", err.Error())
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"clean text", "Hello, world.", "Hello, world."},
		{"newline", "Hello\nworld", "Hello world"},
		{"carriage return", "Hello\rworld", "Hello world"},
		{"crlf", "Hello\r\nworld", "Hello world"},
		{"multiple newlines", "Hello\n\n\nworld", "Hello world"},
		{"tabs", "Hello\tworld", "Hello world"},
		{"mixed control chars", "Hello\r\n\tworld\n", "Hello world"},
		{"DEL character", "Hello\x7Fworld", "Hello world"},
		{"null byte", "Hello\x00world", "Hello world"},
		{"unicode preserved", "caf\u00e9 na\u00efve \u00fcber", "caf\u00e9 na\u00efve \u00fcber"},
		{"empty input", "", ""},
		{"only whitespace", "  \n\r\t  ", ""},
		{"leading trailing newlines", "\nHello world\n", "Hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitize(tt.input)
			if got != tt.want {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestClientTranscribeStripsNewlines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := inferenceResponse{Text: "Hello\nworld\r\nfoo"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	client := NewClient(addr)

	text, err := client.Transcribe([]byte("fake wav"), "")
	if err != nil {
		t.Fatalf("transcribe failed: %v", err)
	}
	if text != "Hello world foo" {
		t.Errorf("expected 'Hello world foo', got %q", text)
	}
}

func TestServerWaitReady(t *testing.T) {
	// Start a TCP listener to simulate a ready server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	addr := ln.Addr().String()
	host, port, _ := net.SplitHostPort(addr)
	var portNum int
	fmt.Sscanf(port, "%d", &portNum)

	srv := NewServer(ServerConfig{
		Host: host,
		Port: portNum,
	})

	ctx := context.Background()
	err = srv.waitReady(ctx, 2*time.Second)
	if err != nil {
		t.Fatalf("waitReady should succeed with listener: %v", err)
	}
}

func TestServerWaitReadyTimeout(t *testing.T) {
	// Use a port with no listener
	srv := NewServer(ServerConfig{
		Host: "127.0.0.1",
		Port: 19999,
	})

	ctx := context.Background()
	err := srv.waitReady(ctx, 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout in error, got %q", err.Error())
	}
}

func TestServerWaitReadyCancelled(t *testing.T) {
	srv := NewServer(ServerConfig{
		Host: "127.0.0.1",
		Port: 19998,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := srv.waitReady(ctx, 5*time.Second)
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
}
