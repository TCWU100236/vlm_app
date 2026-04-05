package main

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

var llm *ollama.LLM

// 透過 langchaingo 做 VLM inference
func vlmInference(prompt string, jpegBytes []byte) (string, error) {
	if prompt == "" {
		prompt = "你在畫面中看到哪些東西和資訊？"
	}

	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.BinaryPart("image/jpeg", jpegBytes),
				llms.TextPart(prompt),
			},
		}},
	)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Content, nil
}

func testHandler(c *gin.Context) {
	prompt := c.PostForm("prompt")   // 讀取文字欄位
	file, err := c.FormFile("image") // 讀取檔案欄位
	if err != nil {
		c.JSON(400, gin.H{"error": "圖片是必填"})
		return
	}

	// 開啟檔案讀取內容
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	// img, _ := os.ReadFile("cat.jpg")
	img, _ := io.ReadAll(src)

	// 呼叫 VLM 推論
	resp, err := vlmInference(prompt, img)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"result": resp})
}

func main() {
	// VLM 初始化
	var err error
	llm, err = ollama.New(ollama.WithModel("qwen3-vl:2b"))
	if err != nil {
		panic("Ollama 連線失敗: " + err.Error())
	}

	r := gin.Default() // 建立含 Logger + Recovery 的 engine
	r.POST("/test", testHandler)
	r.Run(":8081") // 啟動在 port 8080
}
