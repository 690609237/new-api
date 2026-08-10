package service

import (
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"

	"github.com/stretchr/testify/assert"
)

func TestExtractLatestUserMessageExcludesContext(t *testing.T) {
	tests := []struct {
		name    string
		request dto.Request
		want    string
	}{
		{
			name: "openai chat",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{
				{Role: "system", Content: "system prompt"},
				{Role: "user", Content: "older user message"},
				{Role: "assistant", Content: "older assistant response"},
				{Role: "user", Content: []any{
					map[string]any{"type": "text", "text": "latest line one"},
					map[string]any{"type": "image_url", "image_url": "data:image/png;base64,secret"},
					map[string]any{"type": "text", "text": "latest line two"},
				}},
			}},
			want: "latest line one\nlatest line two",
		},
		{
			name: "claude messages",
			request: &dto.ClaudeRequest{Messages: []dto.ClaudeMessage{
				{Role: "user", Content: "old"},
				{Role: "assistant", Content: "reply"},
				{Role: "user", Content: []any{
					map[string]any{"type": "text", "text": "latest claude"},
					map[string]any{"type": "image", "source": map[string]any{"data": "secret"}},
				}},
			}},
			want: "latest claude",
		},
		{
			name: "gemini contents",
			request: &dto.GeminiChatRequest{Contents: []dto.GeminiChatContent{
				{Role: "user", Parts: []dto.GeminiPart{{Text: "old"}}},
				{Role: "model", Parts: []dto.GeminiPart{{Text: "reply"}}},
				{Role: "user", Parts: []dto.GeminiPart{{Text: "latest gemini"}, {Text: "hidden thought", Thought: true}}},
			}},
			want: "latest gemini",
		},
		{
			name: "image prompt",
			request: &dto.ImageRequest{
				Prompt: "draw a lighthouse",
			},
			want: "draw a lighthouse",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, ExtractLatestUserMessage(test.request))
		})
	}
}

func TestExtractLatestUserMessageFromResponsesInput(t *testing.T) {
	input := []byte(`[
		{"role":"user","content":[{"type":"input_text","text":"old"}]},
		{"role":"assistant","content":[{"type":"output_text","text":"reply"}]},
		{"role":"user","content":[
			{"type":"input_text","text":"latest one"},
			{"type":"input_image","image_url":"data:image/png;base64,secret"},
			{"type":"input_text","text":"latest two"}
		]},
		{"type":"function_call_output","content":"tool output must not replace the user message"}
	]`)
	request := &dto.OpenAIResponsesRequest{Input: input}

	assert.Equal(t, "latest one\nlatest two", ExtractLatestUserMessage(request))
}
