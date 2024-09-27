package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

const (
	sampleInterval = 1 * time.Second
)

type sample struct {
	memory    int
	timestamp time.Time
}

type ChartFrame struct {
	content *fyne.Container
	u       *UI
}

func NewChartFrame(u *UI) *ChartFrame {
	f := &ChartFrame{
		content: container.NewCenter(widget.NewLabel("Select a process to start")),
		u:       u,
	}
	return f
}

func (f *ChartFrame) Start(pid int32) error {
	vv := make([]sample, 0)
	f.content.RemoveAll()
	ticker := time.NewTicker(sampleInterval)
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	name, err := p.Name()
	if err != nil {
		return err
	}
	go func() {
		for {
			f.content.RemoveAll()
			func() {
				stats, err := p.MemoryInfoEx()
				if err != nil {
					log.Fatal(err)
				}
				vv = append(vv, sample{
					memory:    int(stats.RSS - stats.Shared),
					timestamp: time.Now(),
				})
				if len(vv) == 0 {
					f.content.Add(widget.NewLabel("No data"))
					return
				}
				c, err := f.makeChart(name, vv, sampleInterval)
				if err != nil {
					f.content.Add(widget.NewLabel("No data"))
					return
				}
				f.content.Add(c)
			}()
			f.content.Refresh()
			<-ticker.C
		}
	}()
	return nil
}

func (f *ChartFrame) makeChart(name string, d []sample, t time.Duration) (fyne.CanvasObject, error) {
	s := f.u.w.Canvas().Scale()
	w, h := float32(900), float32(450)
	content, err := makeRawChart(name, d, int(s*w), int(s*h))
	if err != nil {
		return nil, err
	}
	r := fyne.NewStaticResource("dummy", content)
	chart := canvas.NewImageFromResource(r)
	chart.FillMode = canvas.ImageFillContain
	chart.SetMinSize(fyne.Size{Width: w, Height: h})
	title := widget.NewRichTextFromMarkdown(
		fmt.Sprintf("## %s - Memory usage over time (T=%v)", name, t),
	)
	return container.NewBorder(
		container.NewHBox(layout.NewSpacer(), title, layout.NewSpacer()),
		nil,
		nil,
		nil,
		chart,
	), nil
}

func makeRawChart(name string, d []sample, w, h int) ([]byte, error) {
	xv := make([]time.Time, len(d))
	yv := make([]float64, len(d))
	for i, v := range d {
		xv[i] = v.timestamp
		yv[i] = float64(v.memory)
	}
	series := []chart.Series{
		chart.TimeSeries{
			Name:    name,
			Style:   chart.StyleTextDefaults(),
			XValues: xv,
			YValues: yv,
		},
	}
	defaultStyle := chart.Style{
		Show:        true,
		FontColor:   chartColorFromFyne(theme.ColorNameForeground),
		StrokeColor: chartColorFromFyne(theme.ColorNameForeground),
		FillColor:   chart.ColorTransparent,
	}
	graph := chart.Chart{
		Width:  w,
		Height: h,
		Background: chart.Style{
			FillColor: chart.ColorTransparent,
			Padding: chart.Box{
				Top:    25,
				Bottom: 25,
			},
		},
		Canvas: chart.Style{
			FillColor: chart.ColorTransparent,
		},
		XAxis: chart.XAxis{
			Name:  "Time",
			Style: defaultStyle,
			ValueFormatter: func(x any) string {
				v := x.(float64)
				t := time.Unix(0, int64(v))
				return humanize.Time(t)
			},
		},
		YAxis: chart.YAxis{
			Name:  "Bytes",
			Style: defaultStyle,
			ValueFormatter: func(x any) string {
				v := x.(float64)
				return humanize.Bytes(uint64(v))
			}},
		Series: series,
	}
	var buf bytes.Buffer
	if err := graph.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func chartColorFromFyne(cn fyne.ThemeColorName) drawing.Color {
	c := theme.Color(cn)
	r, g, b, a := c.RGBA()
	return drawing.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}
