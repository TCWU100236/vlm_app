package ui

import (
	// "log"
	"context"
	"fyneTest/camera"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildUI() {
	// 用來控制 VLM 推論的開關
	isRunning := false
	// 建立可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

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

	// 先宣告按鈕變數，才能在 callback 裡互相控制
	var startBtn, stopBtn, exitBtn *widget.Button

	startBtn = widget.NewButton("開始 inference", func() {
		ctx, cancel = context.WithCancel(context.Background()) // 重新建立一個新的 context
		isRunning = true

		// 更新元件狀態
		startBtn.Disable()
		stopBtn.Enable()
		exitBtn.Disable()
		promptEntry.Disable()
	})

	stopBtn = widget.NewButton("暫停 inference", func() {
		isRunning = false
		cancel()

		// 更新元件狀態
		startBtn.Enable()
		stopBtn.Disable()
		exitBtn.Enable()
		promptEntry.Enable()
	})

	exitBtn = widget.NewButton("結束應用程式", func() {
		a.Quit()
	})

	// 程式剛啟動時的預設狀態
	stopBtn.Disable()
	btnRow := container.NewHBox(startBtn, stopBtn)
	promptEntry.SetText("你看到哪些東西？")

	// 啟動攝影機串流
	go camera.StreamCamera(canvasImg, promptEntry, outputLabel, &isRunning, &ctx) // 丟到背景執行（goroutine）

	// 右側佈局：從上到下
	rightPanel := container.NewVBox(
		widget.NewLabel("Prompt"),
		promptEntry,
		widget.NewSeparator(),
		widget.NewLabel("VLM 輸出"),
		outputLabel,
		widget.NewSeparator(),
		btnRow,
		exitBtn,
	)

	// 左右合併
	split := container.NewHSplit(canvasImg, rightPanel)
	split.Offset = 0.6 // 左邊佔 60%

	w.SetContent(split)
	w.ShowAndRun()
}
