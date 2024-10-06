package ui

import (
	"cmp"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	datatable "github.com/ErikKalkoken/fyne-datatable"
	"github.com/shirou/gopsutil/v4/process"
)

func (u *UI) showSelectProcessModal() {
	f := NewSelectProcessFrame(u)
	d := dialog.NewCustom("Choose process", "Cancel", f.selector, u.w)
	f.dialog = d
	d.Show()
	d.Resize(fyne.Size{Width: 600, Height: 450})
}

type SelectProcessFrame struct {
	dialog   *dialog.CustomDialog
	selector fyne.CanvasObject
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
	pidStr string // TODO: DELETE
	user   string
}

func (e *entry) ToRow() []string {
	r := make([]string, 3)
	r[0] = e.name
	r[1] = e.user
	r[2] = strconv.Itoa(int(e.pid))
	return r
}

const (
	defaultInterval = 0
)

var intervals = []time.Duration{
	1 * time.Second,
	10 * time.Second,
	30 * time.Second,
	60 * time.Second,
}

func (f *SelectProcessFrame) makeSelector() fyne.CanvasObject {
	dt, err := datatable.New(datatable.Config{
		Columns: []datatable.Column{
			{Title: "Name"}, {Title: "User"}, {Title: "PID"},
		},
		FooterHidden: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	entries := fetchProcessList()
	data := make([][]string, len(entries))
	for i, e := range entries {
		data[i] = e.ToRow()
	}
	if err := dt.SetData(data); err != nil {
		log.Fatal(err)
	}
	var options []string
	for _, x := range intervals {
		options = append(options, x.String())
	}
	interval := widget.NewSelect(options, nil)
	interval.Selected = options[defaultInterval]
	dt.OnSelected = func(index int) {
		t := intervals[defaultInterval]
		for i, o := range options {
			if interval.Selected == o {
				t = intervals[i]
			}
		}
		x := entries[index]
		if err := f.u.cf.Start(x.pid, t); err != nil {
			d2 := dialog.NewError(err, f.u.w)
			d2.Show()
			return
		}
		f.dialog.Hide()
	}
	form := widget.NewForm(widget.NewFormItem("Interval", interval))
	return container.NewBorder(nil, form, nil, nil, dt)
}

func fetchProcessList() []entry {
	entries := make([]entry, 0)
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
			name: name,
			pid:  pid,
			user: username,
		})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		return cmp.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
	})
	return entries
}
