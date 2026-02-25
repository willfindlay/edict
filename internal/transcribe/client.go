package transcribe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client sends audio to the whisper-server HTTP API for transcription.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a transcription client pointing at the given server address.
func NewClient(addr string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s", addr),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Ping checks if the whisper-server is reachable.
func (c *Client) Ping() error {
	resp, err := c.httpClient.Get(c.baseURL + "/")
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// inferenceResponse is the JSON structure returned by whisper-server /inference.
type inferenceResponse struct {
	Text string `json:"text"`
}

// Transcribe sends WAV audio data with an optional initial prompt and returns the transcribed text.
func (c *Client) Transcribe(wavData []byte, prompt string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the audio file
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(wavData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	// Add parameters
	fields := map[string]string{
		"temperature":     "0.0",
		"temperature_inc": "0.2",
		"response_format": "json",
	}
	if prompt != "" {
		fields["prompt"] = prompt
	}

	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			return "", fmt.Errorf("write field %s: %w", k, err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/inference", &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("whisper-server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result inferenceResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return strings.TrimSpace(result.Text), nil
}
