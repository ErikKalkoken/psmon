package main

import (
	"cmp"
	"log"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v4/process"
)

func (u *UI) showSelectProcessModal() {
	f := NewSelectProcessFrame(u)
	d := dialog.NewCustom("Choose process", "Cancel", f.table, u.w)
	f.dialog = d
	d.Show()
	d.Resize(fyne.Size{Width: 600, Height: 400})
}

type SelectProcessFrame struct {
	dialog *dialog.CustomDialog
	table  *widget.Table
	u      *UI
}

func NewSelectProcessFrame(u *UI) *SelectProcessFrame {
	f := &SelectProcessFrame{
		u: u,
	}
	f.table = f.makeTable()
	return f
}

type entry struct {
	pid  int32
	name string
	exe  string
}

func (f *SelectProcessFrame) makeTable() *widget.Table {
	entries := make([]entry, 0)
	table := widget.NewTableWithHeaders(
		func() (int, int) {
			return len(entries), 3
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("dummy")
		},
		func(tci widget.TableCellID, co fyne.CanvasObject) {
			if tci.Row >= len(entries) {
				return
			}
			cell := co.(*widget.Label)
			x := entries[tci.Row]
			switch tci.Col {
			case 0:
				cell.SetText(strconv.Itoa(int(x.pid)))
			case 1:
				cell.SetText(x.name)
			case 2:
				cell.SetText(x.exe)
			}
		},
	)
	table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("header")
	}
	table.UpdateHeader = func(tci widget.TableCellID, co fyne.CanvasObject) {
		l := co.(*widget.Label)
		var t string
		switch tci.Col {
		case 0:
			t = "PID"
		case 1:
			t = "Name"
		case 2:
			t = "Command"
		}
		l.SetText(t)
	}
	table.ShowHeaderColumn = false
	pids, err := process.Pids()
	if err != nil {
		log.Fatal(err)
	}
	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		name, err := p.Name()
		if err != nil {
			continue
		}
		exe, err := p.Exe()
		if err != nil {
			continue
		}
		if name != "evebuddy" {
			continue
		}
		entries = append(entries, entry{pid: pid, name: name, exe: exe})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return cmp.Compare(a.name, b.name)
	})
	table.OnSelected = func(tci widget.TableCellID) {
		defer table.UnselectAll()
		if tci.Row >= len(entries) {
			return
		}
		x := entries[tci.Row]
		if err := f.u.cf.Start(x.pid); err != nil {
			d2 := dialog.NewError(err, f.u.w)
			d2.Show()
			return
		}
		f.dialog.Hide()
	}
	return table
}
