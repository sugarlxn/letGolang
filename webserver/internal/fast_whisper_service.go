package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FastWhisperService 调用 FastWhisper Python 端点实现 SpeechToTextProvider。
type FastWhisperService struct {
	baseURL *url.URL
	client  *http.Client
}

const (
	defaultASRTimeout = 30 * time.Second
	minPCMLength      = 16000 * 2 // 1 秒的 16kHz 16bit PCM 数据
)

// NewFastWhisperService 构造函数，允许自定义 http.Client。
func NewFastWhisperService(rawURL string, client *http.Client) (*FastWhisperService, error) {
	parsed, err := url.Parse(strings.TrimSuffix(rawURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	if client == nil {
		client = &http.Client{Timeout: defaultASRTimeout}
	}

	return &FastWhisperService{baseURL: parsed, client: client}, nil
}

// Ping 检查服务健康状态
func (f *FastWhisperService) Ping(ctx context.Context) error {
	if f == nil {
		return errors.New("nil FastWhisperService receiver")
	}

	// 假设后端有一个健康检查端点（如果没有，可以用 /api/v1/asr 的 OPTIONS 请求或其他方式）
	// 这里简化为直接返回成功，实际生产环境应该实现真实的健康检查
	return nil
}

// TranscribeFile 通过文件上传接口进行语音识别
// audioData: 音频文件的字节数据
// filename: 文件名（例如 "audio.wav"）
func (f *FastWhisperService) TranscribeFile(ctx context.Context, audioData []byte, filename string) (SpeechToTextResponse, error) {
	var empty SpeechToTextResponse
	if f == nil {
		return empty, errors.New("nil FastWhisperService receiver")
	}

	if len(audioData) == 0 {
		return empty, errors.New("audio data is empty")
	}

	// 创建 multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return empty, fmt.Errorf("create form file: %w", err)
	}

	if _, err := part.Write(audioData); err != nil {
		return empty, fmt.Errorf("write audio data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return empty, fmt.Errorf("close multipart writer: %w", err)
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.resolvePath("/api/v1/asr"), body)
	if err != nil {
		return empty, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := f.client.Do(httpReq)
	if err != nil {
		return empty, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("transcribe failed: status %d, body: %s", resp.StatusCode, respBody)
	}

	// 解析 JSON 响应
	var result SpeechToTextResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return empty, fmt.Errorf("parse response json: %w, body: %s", err, respBody)
	}

	return result, nil
}

// TranscribePCM 通过 PCM 接口进行语音识别（适用于 ESP32 等嵌入式设备）
// pcmData: PCM 音频数据（16kHz, 16bit）
func (f *FastWhisperService) TranscribePCM(ctx context.Context, pcmData []byte) (SpeechToTextResponse, error) {
	var empty SpeechToTextResponse
	if f == nil {
		return empty, errors.New("nil FastWhisperService receiver")
	}

	if len(pcmData) < minPCMLength {
		return SpeechToTextResponse{
			Code:    1,
			Message: "audio too short",
		}, nil
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.resolvePath("/api/v1/asr/pcm"), bytes.NewReader(pcmData))
	if err != nil {
		return empty, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	// 发送请求
	resp, err := f.client.Do(httpReq)
	if err != nil {
		return empty, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return empty, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("transcribe pcm failed: status %d, body: %s", resp.StatusCode, respBody)
	}

	// 解析 JSON 响应
	var result SpeechToTextResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return empty, fmt.Errorf("parse response json: %w, body: %s", err, respBody)
	}

	return result, nil
}

func (f *FastWhisperService) resolvePath(path string) string {
	u := *f.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String()
}
