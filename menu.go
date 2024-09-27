package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (u *UI) makeMenu() *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Select Process...", func() {
			u.showSelectProcessModal()
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About...", func() {
			d := dialog.NewInformation("About", "TBD", u.w)
			d.Show()
		}),
	)
	main := fyne.NewMainMenu(fileMenu, helpMenu)
	return main
}
