package service

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"

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
		{
			name: "openai assistant continuation",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{
				{Role: "system", Content: "system prompt"},
				{Role: "user", Content: "already logged user message"},
				{Role: "assistant", Content: "latest assistant response"},
			}},
			want: "",
		},
		{
			name: "openai tool continuation",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{
				{Role: "user", Content: "already logged user message"},
				{Role: "assistant", Content: "calling a tool"},
				{Role: "tool", Content: "tool output"},
			}},
			want: "",
		},
		{
			name: "openai system only",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{
				{Role: "system", Content: "must never be logged"},
			}},
			want: "",
		},
		{
			name: "claude assistant continuation",
			request: &dto.ClaudeRequest{Messages: []dto.ClaudeMessage{
				{Role: "user", Content: "already logged user message"},
				{Role: "assistant", Content: "latest assistant response"},
			}},
			want: "",
		},
		{
			name: "gemini model continuation",
			request: &dto.GeminiChatRequest{Contents: []dto.GeminiChatContent{
				{Role: "user", Parts: []dto.GeminiPart{{Text: "already logged user message"}}},
				{Role: "model", Parts: []dto.GeminiPart{{Text: "latest model response"}}},
			}},
			want: "",
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
		]}
	]`)
	request := &dto.OpenAIResponsesRequest{Input: input}

	assert.Equal(t, "latest one\nlatest two", ExtractLatestUserMessage(request))
}

func TestExtractLatestUserMessageExcludesResponsesContinuation(t *testing.T) {
	input := []byte(`[
		{"role":"user","content":[{"type":"input_text","text":"already logged user message"}]},
		{"type":"function_call","name":"read_file","arguments":"{}"},
		{"type":"function_call_output","content":"tool output must not be logged"}
	]`)
	request := &dto.OpenAIResponsesRequest{
		Input:        input,
		Instructions: []byte(`"system instructions must not be logged"`),
	}

	assert.Empty(t, ExtractLatestUserMessage(request))
}

func TestExtractLatestUserMessageFromResponsesTopLevelInputText(t *testing.T) {
	request := &dto.OpenAIResponsesRequest{
		Input: []byte(`[{"type":"input_text","text":"fresh user input"}]`),
	}

	assert.Equal(t, "fresh user input", ExtractLatestUserMessage(request))
}

func TestExtractUserMessageForLogExcludesAutomatedCodexRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name           string
		subagentHeader string
		memgenHeader   string
		turnHeader     string
		clientMetadata []byte
		request        dto.Request
	}{
		{
			name:           "subagent header",
			subagentHeader: "collab_spawn",
		},
		{
			name:       "compaction header",
			turnHeader: `{"request_kind":"compaction"}`,
		},
		{
			name:         "memory generation header",
			memgenHeader: "1",
		},
		{
			name:           "subagent client metadata",
			clientMetadata: []byte(`{"x-openai-subagent":"guardian"}`),
		},
		{
			name:           "encoded turn metadata",
			clientMetadata: []byte(`{"x-codex-turn-metadata":"{\"request_kind\":\"turn\",\"subagent_kind\":\"thread_spawn\"}"}`),
		},
		{
			name:           "automation turn",
			clientMetadata: []byte(`{"x-codex-turn-metadata":"{\"request_kind\":\"turn\",\"thread_source\":\"automation\"}"}`),
		},
		{
			name: "title generation prompt",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{{
				Role:    "user",
				Content: "You are a helpful assistant. You will be presented with a user prompt, and your job is to provide a short title for a task that will be created from that prompt.\n\nUser prompt:\ncheck my plan",
			}}},
		},
		{
			name: "suggestion generation prompt",
			request: &dto.GeneralOpenAIRequest{Messages: []dto.Message{{
				Role:    "user",
				Content: "# Overview\n\nGenerate 0 to 3 hyperpersonalized suggestions for what this user can do with Codex in this local project: /workspace/project\n\n# Rules",
			}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			ctx.Request.Header.Set("x-openai-subagent", test.subagentHeader)
			ctx.Request.Header.Set("x-openai-memgen-request", test.memgenHeader)
			ctx.Request.Header.Set("x-codex-turn-metadata", test.turnHeader)
			request := test.request
			if request == nil {
				request = &dto.OpenAIResponsesRequest{
					Input:          []byte(`"system-generated text presented as user input"`),
					ClientMetadata: test.clientMetadata,
				}
			}

			assert.Empty(t, ExtractUserMessageForLog(ctx, request))
		})
	}
}

func TestExtractUserMessageForLogKeepsDirectCodexUserTurn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	request := &dto.OpenAIResponsesRequest{
		Input: []byte(`"direct user input"`),
		ClientMetadata: []byte(`{
			"x-codex-turn-metadata":"{\"request_kind\":\"turn\"}"
		}`),
	}

	assert.Equal(t, "direct user input", ExtractUserMessageForLog(ctx, request))
}
