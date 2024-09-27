package main

import (
	"cmp"
	"log"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v4/process"
)

var intervals = []time.Duration{
	1 * time.Second,
	10 * time.Second,
	30 * time.Second,
	60 * time.Second,
}

func (u *UI) showSelectProcessModal() {
	f := NewSelectProcessFrame(u)
	d := dialog.NewCustom("Choose process", "Cancel", f.selector, u.w)
	f.dialog = d
	d.Show()
	d.Resize(fyne.Size{Width: 600, Height: 450})
}

type SelectProcessFrame struct {
	dialog   *dialog.CustomDialog
	selector *fyne.Container
	u        *UI
}

func NewSelectProcessFrame(u *UI) *SelectProcessFrame {
	f := &SelectProcessFrame{
		u: u,
	}
	f.selector = f.makeSelector()
	return f
}

type entry struct {
	name   string
	pid    int32
	pidStr string
	user   string
}

func (f *SelectProcessFrame) makeSelector() *fyne.Container {
	entries := make([]entry, 0)
	var selection []entry
	var colLayout columnsLayout
	list := widget.NewList(
		func() int {
			return len(selection)
		},
		func() fyne.CanvasObject {
			return container.New(
				colLayout,
				widget.NewLabel("dummy"),
				widget.NewLabel("dummy"),
				widget.NewLabel("dummy"),
			)
		},
		func(id widget.ListItemID, co fyne.CanvasObject) {
			if id >= len(selection) {
				return
			}
			x := selection[id]
			c := co.(*fyne.Container)
			l1 := c.Objects[0].(*widget.Label)
			l1.SetText(x.name)
			l2 := c.Objects[1].(*widget.Label)
			l2.SetText(x.user)
			l3 := c.Objects[2].(*widget.Label)
			l3.SetText(x.pidStr)
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
		username, err := p.Username()
		if err != nil {
			continue
		}
		entries = append(entries, entry{
			name:   name,
			pid:    pid,
			pidStr: strconv.Itoa(int(pid)),
			user:   username,
		})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return cmp.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	widths := make([]float32, 3)
	for _, x := range entries {
		l1 := widget.NewLabel(x.name)
		w1 := l1.MinSize().Width
		if widths[0] < w1 {
			widths[0] = w1
		}
		l2 := widget.NewLabel(x.user)
		w2 := l2.MinSize().Width
		if widths[1] < w2 {
			widths[1] = w2
		}
		l3 := widget.NewLabel(x.pidStr)
		w3 := l3.MinSize().Width
		if widths[2] < w3 {
			widths[2] = w3
		}
	}
	widths2 := make([]int, 3)
	for i, v := range widths {
		widths2[i] = int(math.Ceil(float64(v)))
	}
	colLayout = columnsLayout(widths2)
	selection = slices.Clone(entries)
	var options []string
	for _, x := range intervals {
		options = append(options, x.String())
	}
	interval := widget.NewSelect(options, nil)
	interval.Selected = options[2]
	list.OnSelected = func(id widget.ListItemID) {
		defer list.UnselectAll()
		if id >= len(selection) {
			return
		}
		t := intervals[2]
		for i, o := range options {
			if interval.Selected == o {
				t = intervals[i]
			}
		}
		x := selection[id]
		if err := f.u.cf.Start(x.pid, t); err != nil {
			d2 := dialog.NewError(err, f.u.w)
			d2.Show()
			return
		}
		f.dialog.Hide()
	}
	e1 := widget.NewEntry()
	e1.PlaceHolder = "Name"
	e2 := widget.NewEntry()
	e2.PlaceHolder = "User"
	e3 := widget.NewEntry()
	e3.PlaceHolder = "ID"
	filterEntries := func(s1, s2, s3 string) {
		var x []entry
		for _, e := range entries {
			if (s1 == "" || strings.Contains(e.name, s1)) &&
				(s2 == "" || strings.Contains(e.user, s2)) &&
				(s3 == "" || strings.Contains(e.pidStr, s3)) {
				x = append(x, e)
			}
		}
		selection = x
		list.Refresh()
	}
	e1.OnChanged = func(s string) {
		filterEntries(s, e2.Text, e3.Text)
	}
	e2.OnChanged = func(s string) {
		filterEntries(e1.Text, s, e3.Text)
	}
	e3.OnChanged = func(s string) {
		filterEntries(e1.Text, e2.Text, s)
	}
	header := container.New(colLayout, e1, e2, e3)
	form := widget.NewForm(widget.NewFormItem("Interval", interval))
	return container.NewBorder(header, form, nil, nil, list)
}

type columnsLayout []int

func (d columnsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := float32(0), float32(0)
	if len(objects) > 0 {
		h = objects[0].MinSize().Height
	}
	for i, x := range d {
		w += float32(x)
		if i < len(d)-1 {
			w += theme.Padding()
		}
	}
	return fyne.NewSize(w, h)
}

func (d columnsLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	pos := fyne.NewPos(0, 0)
	var lastX float32
	for i, o := range objects {
		size := o.MinSize()
		o.Resize(fyne.Size{Width: float32(d[i]), Height: size.Height})
		o.Move(pos)
		var x float32
		if len(d) > i {
			x = float32(d[i])
			lastX = x
		} else {
			x = lastX
		}
		pos = pos.AddXY(x+theme.Padding(), 0)
	}
}
