package ui

import (
	// "log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func BuildUI() {
	a := app.New()
	w := a.NewWindow("VLM Demo")
	w.Resize(fyne.NewSize(1000, 600))

	// 左側：攝影機畫面
	cameraImg := widget.NewLabel("VLM 輸出將顯示在這裡")

	// 右側元件
	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetPlaceHolder("輸入你的 prompt...")
	promptEntry.SetMinRowsVisible(4)

	outputLabel := widget.NewLabel("VLM 輸出將顯示在這裡")
	outputLabel.Wrapping = fyne.TextWrapWord

	startBtn := widget.NewButton("開始 inference", func() {
		// TODO: 更新btn狀態
	})

	stopBtn := widget.NewButton("暫停 inference", func() {
		// TODO: 更新元件狀態
	})

	exitBtn := widget.NewButton("結束應用程式", func() {
		a.Quit()
	})

	// 程式剛啟動時的預設狀態
	stopBtn.Disable()
	btnRow := container.NewHBox(startBtn, stopBtn)

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
	split := container.NewHSplit(cameraImg, rightPanel)
	split.Offset = 0.6 // 左邊佔 60%

	w.SetContent(split)
	w.ShowAndRun()
}
