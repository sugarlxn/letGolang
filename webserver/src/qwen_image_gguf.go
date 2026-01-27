package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// QwenImageGGUF 调用 Python 端点实现 TextToImageProvider。
type QwenImageGGUF struct {
	baseURL *url.URL
	client  *http.Client
}

const (
	defaultGenerateTimeout = 60 * time.Second
	defaultInferenceSteps  = 20
)

// NewQwenImageGGUF 构造函数，允许自定义 http.Client。
func NewQwenImageGGUF(rawURL string, client *http.Client) (*QwenImageGGUF, error) {
	parsed, err := url.Parse(strings.TrimSuffix(rawURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	if client == nil {
		client = &http.Client{Timeout: defaultGenerateTimeout}
	}

	return &QwenImageGGUF{baseURL: parsed, client: client}, nil
}

func (q *QwenImageGGUF) Ping(ctx context.Context) error {
	if q == nil {
		return errors.New("nil QwenImageGGUF receiver")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, q.resolvePath("/ping"), nil)
	if err != nil {
		return fmt.Errorf("create ping request: %w", err)
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("execute ping request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("ping failed: status %d, body: %s", resp.StatusCode, detail)
	}

	var payload struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode ping response: %w", err)
	}

	if strings.ToLower(payload.Status) != "ready" {
		return fmt.Errorf("model not ready: %s", payload.Status)
	}

	return nil
}

func (q *QwenImageGGUF) Generate(ctx context.Context, req TextToImageRequest) (TextToImageResponse, error) {
	var empty TextToImageResponse
	if q == nil {
		return empty, errors.New("nil QwenImageGGUF receiver")
	}

	steps := req.Steps
	if steps <= 0 {
		steps = defaultInferenceSteps
	}

	payload := struct {
		Prompt            string `json:"prompt"`
		NegativePrompt    string `json:"negative_prompt,omitempty"`
		NumInferenceSteps int    `json:"num_inference_steps"`
	}{
		Prompt:            req.Prompt,
		NegativePrompt:    req.NegativePrompt,
		NumInferenceSteps: steps,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return empty, fmt.Errorf("marshal generate payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, q.resolvePath("/generate"), bytes.NewReader(body))
	if err != nil {
		return empty, fmt.Errorf("create generate request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := q.client.Do(httpReq)
	if err != nil {
		return empty, fmt.Errorf("execute generate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return empty, fmt.Errorf("generate failed: status %d, body: %s", resp.StatusCode, detail)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, fmt.Errorf("read image payload: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	return TextToImageResponse{ImageData: imageData, MimeType: mimeType}, nil
}

func (q *QwenImageGGUF) resolvePath(path string) string {
	u := *q.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String()
}
