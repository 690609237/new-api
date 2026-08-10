package service

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
)

// ExtractLatestUserMessage returns only the newest user-supplied text from a
// validated relay request. Historical messages, system prompts, tool calls,
// files, images, and audio payloads are deliberately excluded.
func ExtractLatestUserMessage(request dto.Request) string {
	switch req := request.(type) {
	case *dto.GeneralOpenAIRequest:
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if !strings.EqualFold(req.Messages[i].Role, "user") {
				continue
			}
			parts := make([]string, 0)
			for _, part := range req.Messages[i].ParseContent() {
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
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if !strings.EqualFold(req.Messages[i].Role, "user") {
				continue
			}
			if req.Messages[i].IsStringContent() {
				return req.Messages[i].GetStringContent()
			}
			content, err := req.Messages[i].ParseContent()
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

func latestGeminiUserMessage(request *dto.GeminiChatRequest) string {
	if len(request.Contents) == 0 && len(request.Requests) > 0 {
		return latestGeminiUserMessage(&request.Requests[len(request.Requests)-1])
	}
	for i := len(request.Contents) - 1; i >= 0; i-- {
		content := request.Contents[i]
		if content.Role != "" && !strings.EqualFold(content.Role, "user") {
			continue
		}
		parts := make([]string, 0)
		for _, part := range content.Parts {
			if part.Text != "" && !part.Thought {
				parts = append(parts, part.Text)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
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

	var inputs []dto.Input
	if err := common.Unmarshal(input, &inputs); err != nil {
		return ""
	}
	for i := len(inputs) - 1; i >= 0; i-- {
		if !strings.EqualFold(inputs[i].Role, "user") {
			continue
		}
		if common.GetJsonType(inputs[i].Content) == "string" {
			var text string
			if err := common.Unmarshal(inputs[i].Content, &text); err == nil {
				return text
			}
			continue
		}
		var parts []dto.MediaInput
		if err := common.Unmarshal(inputs[i].Content, &parts); err != nil {
			continue
		}
		texts := make([]string, 0)
		for _, part := range parts {
			if part.Type == "input_text" && part.Text != "" {
				texts = append(texts, part.Text)
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}
