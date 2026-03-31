package vlm

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

var llm, err = ollama.New(ollama.WithModel("qwen3-vl:2b"))

// VLM 做推論
func VLM_inference(ctx context.Context, imageBytes []byte, prompt string) (string, error) {
	// llm, err := ollama.New(ollama.WithModel("qwen3-vl:2b"))
	// if err != nil {
	// 	log.Fatal("無法連接 Ollama:", err)
	// }

	// 送出圖片 + 問題
	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.BinaryPart("image/jpeg", imageBytes),
					llms.TextPart("請用繁體中文回答以下問題\n" + prompt),
				},
			},
		},
	)

	// 加上這兩個檢查
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", nil
	}

	fmt.Println(resp.Choices[0].Content)
	return resp.Choices[0].Content, nil
}
