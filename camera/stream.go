package camera

import (
	"fmt"

	"fyne.io/fyne/v2/canvas"
	"gocv.io/x/gocv"
)

func StreamCamera(canvasImg *canvas.Image) {
	// 開啟攝影機（0 代表第一台）
	cam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Println("無法開啟攝影機:", err)
		return
	}
	defer cam.Close() // 確保在函式結束時釋放攝影機資源

	/*
		一張圖片在電腦裡本質上就是一個矩陣，每個格子存一個像素的顏色值
		每個像素由 3 個數值組成（RGB 格式）
	*/
	frame := gocv.NewMat() // 建立一個空的 Mat 來存影像
	defer frame.Close()

	// 建立顯示視窗
	// window := gocv.NewWindow("攝影機畫面")
	// defer window.Close()

	// 無限迴圈，持續讀取、顯示畫面
	for {
		ok := cam.Read(&frame) // 讀一幀，讀取成功會回傳 true，失敗則回傳 false

		if ok {
			// Mat 轉成 image.Image
			img, err := frame.ToImage()
			if err != nil {
				fmt.Println("轉換失敗:", err)
				continue
			}

			// 替換圖片並通知重繪
			canvasImg.Image = img
			canvasImg.Refresh()
		}
	}
}
