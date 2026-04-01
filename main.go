package main

import (
	"fmt"

	"fyneTest/ui"
	"fyneTest/vlm"

	"gocv.io/x/gocv"
)

func main() {
	ui.BuildUI()

	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Println("錯誤, 無法打開鏡頭!")
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("Qwen-VLM window")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	// 用 channel 傳圖片給 VLM goroutine
	imgChan := make(chan []byte, 1) // buffer=1，避免阻塞

	// VLM 獨立跑在背景
	go func() {
		for imageBytes := range imgChan {
			vlm.VLM_output(imageBytes)
		}
	}()

	frameCount := 0
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Println("無法讀取影像")
			continue
		}
		if img.Empty() {
			continue
		}

		// 每 30 幀才送一次給 VLM（避免洗爆 API）
		frameCount++
		if frameCount%30 == 0 {
			buf, err := gocv.IMEncode(".jpg", img)
			if err == nil {
				imageBytes := buf.GetBytes()
				// 複製一份，避免 buf 釋放後資料消失
				copyBytes := make([]byte, len(imageBytes))
				copy(copyBytes, imageBytes)

				// non-blocking 送出，如果 VLM 還在忙就跳過
				select {
				case imgChan <- copyBytes:
				default:
				}
				buf.Close() // 記得釋放
			}
		}

		window.IMShow(img)
		if window.WaitKey(1) == 27 { // ESC 鍵離開
			break
		}
	}
	close(imgChan)
}
