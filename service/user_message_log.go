package service

import (
	"encoding/json"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	"github.com/gin-gonic/gin"
)

// ExtractUserMessageForLog filters automated Codex requests before extracting
// the current user turn. Codex marks subagents, compaction, prewarm, memory, and
// automation turns in headers and Responses client_metadata.
func ExtractUserMessageForLog(c *gin.Context, request dto.Request) string {
	if c != nil {
		if strings.TrimSpace(c.GetHeader("x-openai-subagent")) != "" {
			return ""
		}
		if strings.TrimSpace(c.GetHeader("x-openai-memgen-request")) != "" {
			return ""
		}
		if isAutomatedCodexTurnMetadata([]byte(c.GetHeader("x-codex-turn-metadata"))) {
			return ""
		}
	}

	if req, ok := request.(*dto.OpenAIResponsesRequest); ok && len(req.ClientMetadata) > 0 {
		var metadata struct {
			Subagent     string          `json:"x-openai-subagent"`
			TurnMetadata json.RawMessage `json:"x-codex-turn-metadata"`
		}
		if err := common.Unmarshal(req.ClientMetadata, &metadata); err == nil {
			if strings.TrimSpace(metadata.Subagent) != "" {
				return ""
			}
			turnMetadata := metadata.TurnMetadata
			if common.GetJsonType(turnMetadata) == "string" {
				var encoded string
				if err := common.Unmarshal(turnMetadata, &encoded); err == nil {
					turnMetadata = []byte(encoded)
				}
			}
			if isAutomatedCodexTurnMetadata(turnMetadata) {
				return ""
			}
		}
	}

	message := ExtractLatestUserMessage(request)
	if isAutomatedCodexPrompt(message) {
		return ""
	}
	return message
}

func isAutomatedCodexPrompt(message string) bool {
	message = strings.TrimSpace(message)
	automatedPromptPrefixes := [...]string{
		"You are a helpful assistant. You will be presented with a user prompt, and your job is to provide a short title for a task that will be created from that prompt.",
		"# Overview\n\nGenerate 0 to 3 hyperpersonalized suggestions for what this user can do with Codex in this local project:",
	}
	for _, prefix := range automatedPromptPrefixes {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

func isAutomatedCodexTurnMetadata(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	var metadata struct {
		RequestKind  string `json:"request_kind"`
		SubagentKind string `json:"subagent_kind"`
		ThreadSource string `json:"thread_source"`
	}
	if err := common.Unmarshal(data, &metadata); err != nil {
		return false
	}
	if metadata.SubagentKind != "" {
		return true
	}
	if metadata.RequestKind != "" && !strings.EqualFold(metadata.RequestKind, "turn") {
		return true
	}
	return strings.EqualFold(metadata.ThreadSource, "automation") ||
		strings.EqualFold(metadata.ThreadSource, "subagent")
}

// ExtractLatestUserMessage returns text only when the current request ends in
// a user-authored turn. Historical messages, system prompts, assistant/tool
// continuations, files, images, and audio payloads are deliberately excluded.
func ExtractLatestUserMessage(request dto.Request) string {
	switch req := request.(type) {
	case *dto.GeneralOpenAIRequest:
		if len(req.Messages) > 0 {
			message := req.Messages[len(req.Messages)-1]
			if !strings.EqualFold(message.Role, "user") {
				return ""
			}
			parts := make([]string, 0)
			for _, part := range message.ParseContent() {
				if part.Type == dto.ContentTypeText && part.Text != "" {
					parts = append(parts, part.Text)
				}
			}
			return strings.Join(parts, "\n")
		}
		if inputs := req.ParseInput(); len(inputs) > 0 {
			return inputs[len(inputs)-1]
		}
		switch prompt := req.Prompt.(type) {
		case string:
			return prompt
		case []string:
			if len(prompt) > 0 {
				return prompt[len(prompt)-1]
			}
		case []any:
			for i := len(prompt) - 1; i >= 0; i-- {
				if text, ok := prompt[i].(string); ok {
					return text
				}
			}
		}
		return ""

	case *dto.ClaudeRequest:
		if len(req.Messages) > 0 {
			message := req.Messages[len(req.Messages)-1]
			if !strings.EqualFold(message.Role, "user") {
				return ""
			}
			if message.IsStringContent() {
				return message.GetStringContent()
			}
			content, err := message.ParseContent()
			if err != nil {
				return ""
			}
			parts := make([]string, 0)
			for _, part := range content {
				if part.Type == "text" && part.GetText() != "" {
					parts = append(parts, part.GetText())
				}
			}
			return strings.Join(parts, "\n")
		}
		return req.Prompt

	case *dto.GeminiChatRequest:
		return latestGeminiUserMessage(req)

	case *dto.OpenAIResponsesRequest:
		return latestResponsesUserMessage(req.Input)

	case *dto.ImageRequest:
		return req.Prompt

	case *dto.AudioRequest:
		return req.Input

	case *dto.EmbeddingRequest:
		if inputs := req.ParseInput(); len(inputs) > 0 {
			return inputs[len(inputs)-1]
		}

	case *dto.RerankRequest:
		return req.Query
	}
	return ""
}

// ExtractLatestUserMessageForModeration is tolerant of Responses API
// continuations. A Responses input may end in a tool output or reasoning item
// without a user role; moderation should still inspect the most recent user
// text earlier in the input sequence.
func ExtractLatestUserMessageForModeration(request dto.Request) string {
	if req, ok := request.(*dto.OpenAIResponsesRequest); ok {
		return latestResponsesUserMessageForModeration(req.Input)
	}
	return ExtractLatestUserMessage(request)
}

func latestGeminiUserMessage(request *dto.GeminiChatRequest) string {
	if len(request.Contents) == 0 && len(request.Requests) > 0 {
		return latestGeminiUserMessage(&request.Requests[len(request.Requests)-1])
	}
	if len(request.Contents) == 0 {
		return ""
	}
	content := request.Contents[len(request.Contents)-1]
	if content.Role != "" && !strings.EqualFold(content.Role, "user") {
		return ""
	}
	parts := make([]string, 0)
	for _, part := range content.Parts {
		if part.Text != "" && !part.Thought {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func latestResponsesUserMessage(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	if common.GetJsonType(input) == "string" {
		var text string
		if err := common.Unmarshal(input, &text); err == nil {
			return text
		}
		return ""
	}

	var inputs []struct {
		Type    string          `json:"type"`
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
		Text    string          `json:"text"`
	}
	if err := common.Unmarshal(input, &inputs); err != nil {
		return ""
	}
	if len(inputs) == 0 {
		return ""
	}
	latest := inputs[len(inputs)-1]
	if latest.Type == "input_text" {
		return latest.Text
	}
	if !strings.EqualFold(latest.Role, "user") {
		return ""
	}
	if common.GetJsonType(latest.Content) == "string" {
		var text string
		if err := common.Unmarshal(latest.Content, &text); err == nil {
			return text
		}
		return ""
	}
	var parts []dto.MediaInput
	if err := common.Unmarshal(latest.Content, &parts); err != nil {
		return ""
	}
	texts := make([]string, 0)
	for _, part := range parts {
		if part.Type == "input_text" && part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func latestResponsesUserMessageForModeration(input []byte) string {
	if len(input) == 0 {
		return ""
	}
	if common.GetJsonType(input) == "string" {
		var text string
		if err := common.Unmarshal(input, &text); err == nil {
			return text
		}
		return ""
	}
	var inputs []struct {
		Type    string          `json:"type"`
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
		Text    string          `json:"text"`
	}
	if err := common.Unmarshal(input, &inputs); err != nil {
		return ""
	}
	for i := len(inputs) - 1; i >= 0; i-- {
		item := inputs[i]
		if item.Type == "input_text" {
			if strings.TrimSpace(item.Text) != "" {
				return item.Text
			}
			continue
		}
		if !strings.EqualFold(item.Role, "user") {
			continue
		}
		if common.GetJsonType(item.Content) == "string" {
			var text string
			if err := common.Unmarshal(item.Content, &text); err == nil && strings.TrimSpace(text) != "" {
				return text
			}
			continue
		}
		var parts []dto.MediaInput
		if err := common.Unmarshal(item.Content, &parts); err != nil {
			continue
		}
		texts := make([]string, 0)
		for _, part := range parts {
			if part.Type == "input_text" && part.Text != "" {
				texts = append(texts, part.Text)
			}
		}
		if text := strings.Join(texts, "\n"); strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}
