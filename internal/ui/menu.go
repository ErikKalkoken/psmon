package ui

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (u *UI) makeMenu() *fyne.MainMenu {
	export := fyne.NewMenuItem("Export...", func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				d2 := dialog.NewError(err, u.w)
				d2.Show()
				return
			}
			if uri == nil {
				return
			}
			name, err := u.cf.Process()
			if err != nil {
				d2 := dialog.NewError(err, u.w)
				d2.Show()
				return
			}
			fn, err := writeCSV(uri, name, u.cf.Samples())
			if err != nil {
				d2 := dialog.NewError(err, u.w)
				d2.Show()
				return
			}
			d2 := dialog.NewInformation("Export completed", fmt.Sprintf("Created file %s", fn), u.w)
			d2.Show()
		}, u.w)
		d.Show()
	})
	export.Disabled = true
	u.fileMenu = fyne.NewMenu("File",
		fyne.NewMenuItem("Open Process...", func() {
			u.showSelectProcessModal()
		}),
		export,
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About...", func() {
			d := dialog.NewInformation("About", "TBD", u.w)
			d.Show()
		}),
	)
	main := fyne.NewMainMenu(u.fileMenu, helpMenu)
	return main
}

func writeCSV(uri fyne.ListableURI, name string, d []sample) (string, error) {
	var records [][]string
	records = append(records, []string{"timestamp", "cpu", "memory"})
	for _, s := range d {
		records = append(records, []string{
			s.timestamp.Format(time.RFC3339),
			fmt.Sprint(s.cpu),
			fmt.Sprint(s.memory),
		})
	}
	start := slices.MinFunc(d, func(a, b sample) int {
		return a.timestamp.Compare(b.timestamp)
	})
	end := slices.MaxFunc(d, func(a, b sample) int {
		return a.timestamp.Compare(b.timestamp)
	})
	l := "20060102T1504"
	fn := fmt.Sprintf("%s_%s-%s.csv", name, start.timestamp.Format(l), end.timestamp.Format(l))
	p := filepath.Join(uri.Path(), fn)
	f, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.WriteAll(records)
	return fn, w.Error()
}
