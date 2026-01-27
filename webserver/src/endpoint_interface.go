package main

import "context"

// 文生图请求结构体
type TextToImageRequest struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	Steps          int    `json:"steps,omitempty"`
	Seed           int64  `json:"seed,omitempty"`
}

// 文生图响应结构体
type TextToImageResponse struct {
	ImageData []byte
	MimeType  string
}

// 语音转文字请求结构体
type SpeechToTextRequest struct {
	Audio []byte `json:"audio"`
}

// ASR 分段结构体
type ASRSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

// 语音转文字响应结构体
type SpeechToTextResponse struct {
	Code                int          `json:"code"`
	Message             string       `json:"message,omitempty"`
	Language            string       `json:"language,omitempty"`
	LanguageProbability float64      `json:"language_probability,omitempty"`
	Segments            []ASRSegment `json:"segments,omitempty"`
}

// 文生图接口，要求后端提供健康检查和生成能力
type TextToImageProvider interface {
	Ping(ctx context.Context) error
	Generate(ctx context.Context, req TextToImageRequest) (TextToImageResponse, error)
}

// 语音转文字接口，要求后端提供健康检查和转录能力
type SpeechToTextProvider interface {
	Ping(ctx context.Context) error
	TranscribeFile(ctx context.Context, audioData []byte, filename string) (SpeechToTextResponse, error)
	TranscribePCM(ctx context.Context, pcmData []byte) (SpeechToTextResponse, error)
}
