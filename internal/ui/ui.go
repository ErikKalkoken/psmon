package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type UI struct {
	a        fyne.App
	cf       *ChartFrame
	w        fyne.Window
	fileMenu *fyne.Menu
}

func NewUI() *UI {
	a := app.NewWithID("psmon.ErikKalkoken.github.com")
	w := a.NewWindow("Process monitor")
	u := &UI{w: w, a: a}
	cf := NewChartFrame(u)
	u.cf = cf
	w.SetContent(cf.content)
	w.SetMainMenu(u.makeMenu())
	w.Resize(fyne.Size{Width: 910, Height: 530})
	return u
}

func (u *UI) ShowAndRun() {
	u.w.ShowAndRun()
}
