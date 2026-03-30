package ui

import (
	// "log"
	"fyneTest/camera"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildUI() {
	isRunning := false
	a := app.New()
	w := a.NewWindow("VLM Camera")
	w.Resize(fyne.NewSize(1000, 600))

	// 左側：攝影機畫面
	canvasImg := canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 640, 480)))
	canvasImg.FillMode = canvas.ImageFillContain

	// 右側元件
	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetPlaceHolder("輸入你的 prompt...")
	promptEntry.SetMinRowsVisible(4)

	outputLabel := widget.NewLabel("VLM 輸出將顯示在這裡")
	outputLabel.Wrapping = fyne.TextWrapWord

	startBtn := widget.NewButton("開始", func() {
		isRunning = true
	})
	stopBtn := widget.NewButton("停止", func() {
		a.Quit()
	})
	btnRow := container.NewHBox(startBtn, stopBtn)

	// 啟動攝影機串流
	go camera.StreamCamera(canvasImg, promptEntry, outputLabel, &isRunning) // 丟到背景執行（goroutine）

	// 右側佈局：從上到下
	rightPanel := container.NewVBox(
		widget.NewLabel("Prompt"),
		promptEntry,
		widget.NewSeparator(),
		widget.NewLabel("VLM 輸出"),
		outputLabel,
		widget.NewSeparator(),
		btnRow,
	)

	// 左右合併
	split := container.NewHSplit(canvasImg, rightPanel)
	split.Offset = 0.6 // 左邊佔 60%

	w.SetContent(split)
	w.ShowAndRun()
}
