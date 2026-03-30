package vlm

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// 在 package 層級建立一次就好
// var llm *ollama.LLM

// 初始化 model
// func Init() {
// 	var err error
// 	llm, err = ollama.New(ollama.WithModel("qwen3-vl:2b"))
// 	if err != nil {
// 		log.Fatal("無法連接 Ollama:", err)
// 	}
// 	fmt.Println("Ollama 連接成功！")
// }

// VLM 做推論
func VLM_inference(imageBytes []byte, prompt string) (string, error) {
	llm, err := ollama.New(ollama.WithModel("qwen3-vl:2b"))
	if err != nil {
		log.Fatal("無法連接 Ollama:", err)
	}

	// 送出圖片 + 問題
	resp, _ := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.BinaryPart("image/jpeg", imageBytes),
					llms.TextPart(prompt),
				},
			},
		},
	)

	fmt.Println(resp.Choices[0].Content)
	return resp.Choices[0].Content, nil
}
