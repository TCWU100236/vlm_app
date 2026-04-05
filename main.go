// package main

// import (
// 	"fmt"

// 	"fyneTest/ui"
// 	"fyneTest/vlm"

// 	"gocv.io/x/gocv"
// )

// func main() {
// 	ui.BuildUI()

// 	webcam, err := gocv.OpenVideoCapture(0)
// 	if err != nil {
// 		fmt.Println("錯誤, 無法打開鏡頭!")
// 		return
// 	}
// 	defer webcam.Close()

// 	window := gocv.NewWindow("Qwen-VLM window")
// 	defer window.Close()

// 	img := gocv.NewMat()
// 	defer img.Close()

// 	// 用 channel 傳圖片給 VLM goroutine
// 	imgChan := make(chan []byte, 1) // buffer=1，避免阻塞

// 	// VLM 獨立跑在背景
// 	go func() {
// 		for imageBytes := range imgChan {
// 			vlm.VLM_output(imageBytes)
// 		}
// 	}()

// 	frameCount := 0
// 	for {
// 		if ok := webcam.Read(&img); !ok {
// 			fmt.Println("無法讀取影像")
// 			continue
// 		}
// 		if img.Empty() {
// 			continue
// 		}

// 		// 每 30 幀才送一次給 VLM（避免洗爆 API）
// 		frameCount++
// 		if frameCount%30 == 0 {
// 			buf, err := gocv.IMEncode(".jpg", img)
// 			if err == nil {
// 				imageBytes := buf.GetBytes()
// 				// 複製一份，避免 buf 釋放後資料消失
// 				copyBytes := make([]byte, len(imageBytes))
// 				copy(copyBytes, imageBytes)

// 				// non-blocking 送出，如果 VLM 還在忙就跳過
// 				select {
// 				case imgChan <- copyBytes:
// 				default:
// 				}
// 				buf.Close() // 記得釋放
// 			}
// 		}

// 		window.IMShow(img)
// 		if window.WaitKey(1) == 27 { // ESC 鍵離開
// 			break
// 		}
// 	}
// 	close(imgChan)
// }

package main

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"gocv.io/x/gocv"
)

// ── Channels's struct ─────────────────────────────
type frameData struct {
	img  image.Image
	jpeg []byte
}

// ────────────────────────────────────────────
// VLM
// ────────────────────────────────────────────

var llm *ollama.LLM

func vlmInfer(jpegBytes []byte, prompt string) string {
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
		return "❌ 錯誤：" + err.Error()
	}
	return resp.Choices[0].Content
}

// ────────────────────────────────────────────
// gocv Mat → JPEG bytes + image.Image
// 用 IMEncode 取代 pixel-by-pixel，速度快很多
// ────────────────────────────────────────────

func matToFrame(mat gocv.Mat) (image.Image, []byte, error) {
	buf, err := gocv.IMEncode(".jpg", mat)
	if err != nil {
		return nil, nil, err
	}
	defer buf.Close()

	b := make([]byte, len(buf.GetBytes()))
	copy(b, buf.GetBytes())

	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, nil, err
	}
	return img, b, nil
}

// ────────────────────────────────────────────
// main
// ────────────────────────────────────────────

func main() {
	// VLM 初始化
	var err error
	llm, err = ollama.New(ollama.WithModel("qwen3-vl:2b"))
	if err != nil {
		panic("Ollama 連線失敗: " + err.Error())
	}

	// 開啟攝影機
	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		panic("無法開啟鏡頭")
	}
	defer webcam.Close()

	// ── Fyne UI 界面設計──────────────────────────────
	a := app.New()
	w := a.NewWindow("VLM Camera")
	w.Resize(fyne.NewSize(1100, 620))

	camImg := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 640, 480)))
	camImg.FillMode = canvas.ImageFillContain
	camImg.SetMinSize(fyne.NewSize(640, 480))

	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetPlaceHolder("輸入 prompt（空白 = 預設問題）")
	promptEntry.SetMinRowsVisible(3)
	promptEntry.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("⏸ 待機中")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	outputLabel := widget.NewLabel("")
	outputLabel.Wrapping = fyne.TextWrapWord
	outputScroll := container.NewVScroll(outputLabel)
	outputScroll.SetMinSize(fyne.NewSize(0, 380))

	running := false
	inferring := false

	var startBtn, stopBtn *widget.Button

	startBtn = widget.NewButton("▶ 開始推論", func() {
		running = true
		startBtn.Disable()
		stopBtn.Enable()
		statusLabel.SetText("🟢 推論中...")
	})
	startBtn.Importance = widget.HighImportance

	stopBtn = widget.NewButton("⏹ 暫停", func() {
		running = false
		startBtn.Enable()
		stopBtn.Disable()
		statusLabel.SetText("⏸ 已暫停")
	})
	stopBtn.Disable()

	exitBtn := widget.NewButton("結束", func() { a.Quit() })

	rightPanel := container.NewVBox(
		widget.NewLabel("📝 Prompt"),
		promptEntry,
		widget.NewSeparator(),
		container.NewHBox(startBtn, stopBtn, exitBtn),
		statusLabel,
		widget.NewSeparator(),
		widget.NewLabel("🔍 推論結果"),
		outputScroll,
	)

	split := container.NewHSplit(camImg, rightPanel)
	split.Offset = 0.6
	w.SetContent(split)

	// ── Channels ─────────────────────────────
	frameChan := make(chan frameData, 1) // 攝影機幀資料
	vlmInChan := make(chan []byte, 1)    // 送給 VLM 的 JPEG bytes
	resultChan := make(chan string, 1)   // VLM 推論結果

	// ── 攝影機 goroutine ──────────────────────
	// 只負責：Read → IMEncode → 送 channel
	go func() {
		mat := gocv.NewMat()
		defer mat.Close()

		for {
			if ok := webcam.Read(&mat); !ok || mat.Empty() {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			img, jpeg, err := matToFrame(mat)
			if err != nil {
				continue
			}
			select {
			case frameChan <- frameData{img, jpeg}:
			default: // UI 還沒消化完就跳過，不堆積
			}
		}
	}()

	// ── VLM goroutine ─────────────────────────
	// 只負責：收 JPEG → 推論 → 送結果
	go func() {
		for jpeg := range vlmInChan {
			prompt := promptEntry.Text
			result := vlmInfer(jpeg, prompt)
			resultChan <- result
		}
	}()

	// ── 更新 UI goroutine ─────────────────────
	frameCount := 0
	go func() {
		ticker := time.NewTicker(33 * time.Millisecond) // ~30fps
		defer ticker.Stop()

		for range ticker.C {
			// 1. 更新 UI 攝影機畫面
			select {
			case f := <-frameChan:
				fyne.Do(func() { // ← Linux 必要，確保在主線程執行
					camImg.Image = f.img
					canvas.Refresh(camImg)
				})

				// 2. 每 30 幀派送一次給 VLM
				frameCount++
				if running && !inferring && frameCount%30 == 0 {
					select {
					case vlmInChan <- f.jpeg:
						inferring = true
						fyne.Do(func() {
							statusLabel.SetText("⏳ 推論中...")
						})
					default:
					}
				}
			default:
			}

			// 3. 接收 VLM inference 結果
			select {
			case result := <-resultChan:
				inferring = false
				ts := time.Now().Format("15:04:05")
				fyne.Do(func() { // ← 同上
					prev := outputLabel.Text
					newText := "[" + ts + "]\n" + result
					if prev != "" {
						newText += "\n──────────\n" + prev
					}
					outputLabel.SetText(newText)
					outputScroll.ScrollToTop()
					if running {
						statusLabel.SetText("✅ 完成，持續偵測中...")
					}
				})
			default:
			}
		}
	}()

	w.ShowAndRun()
}
