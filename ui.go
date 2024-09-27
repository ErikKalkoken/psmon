package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UI struct {
	w       fyne.Window
	a       fyne.App
	cf      *ChartFrame
	content *fyne.Container
}

func NewUI() *UI {
	a := app.New()
	w := a.NewWindow("Process monitor")
	u := &UI{w: w, a: a}
	cf := NewChartFrame(u)
	u.cf = cf
	c := container.NewBorder(
		container.NewHBox(widget.NewButton("Select", func() {
			u.showSelectProcessModal()
		})),
		nil,
		nil,
		nil,
		container.NewCenter(cf.content))
	u.content = c
	w.SetContent(c)
	w.Resize(fyne.Size{Width: 1000, Height: 600})
	return u
}

func (u *UI) ShowAndRun() {
	u.w.ShowAndRun()
}
