package main

import (
	"cmp"
	"fmt"
	"log"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v4/process"
)

func (u *UI) showSelectProcessModal() {
	f := NewSelectProcessFrame(u)
	d := dialog.NewCustom("Choose process", "Cancel", f.list, u.w)
	f.dialog = d
	d.Show()
	d.Resize(fyne.Size{Width: 600, Height: 400})
}

type SelectProcessFrame struct {
	dialog *dialog.CustomDialog
	list   *widget.List
	u      *UI
}

func NewSelectProcessFrame(u *UI) *SelectProcessFrame {
	f := &SelectProcessFrame{
		u: u,
	}
	f.list = f.makeTable()
	return f
}

type entry struct {
	pid  int32
	name string
	exe  string
}

func (f *SelectProcessFrame) makeTable() *widget.List {
	entries := make([]entry, 0)
	list := widget.NewList(
		func() int {
			return len(entries)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("dummy")
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			if id >= len(entries) {
				return
			}
			l := co.(*widget.Label)
			x := entries[id]
			l.SetText(fmt.Sprintf("%s [%d]", x.name, x.pid))
		},
	)
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
	list.OnSelected = func(id widget.ListItemID) {
		defer list.UnselectAll()
		if id >= len(entries) {
			return
		}
		x := entries[id]
		if err := f.u.cf.Start(x.pid); err != nil {
			d2 := dialog.NewError(err, f.u.w)
			d2.Show()
			return
		}
		f.dialog.Hide()
	}
	return list
}
