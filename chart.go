package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type ChartFrame struct {
	numCores int
	content  *fyne.Container
	closeC   *chan struct{}
	u        *UI
}

func NewChartFrame(u *UI) *ChartFrame {
	c, err := cpu.Counts(true)
	if err != nil {
		log.Fatal(err)
	}
	f := &ChartFrame{
		content:  container.NewCenter(widget.NewLabel("Select a process to start")),
		numCores: c,
		u:        u,
	}
	return f
}

func (f *ChartFrame) Start(pid int32, t time.Duration) error {
	vv := make([]sample, 0)
	f.content.RemoveAll()
	f.content.Refresh()
	ticker := time.NewTicker(t)
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	name, err := p.Name()
	if err != nil {
		return err
	}
	if f.closeC != nil {
		*f.closeC <- struct{}{}
	}
	closeC := make(chan struct{})
	f.closeC = &closeC
	go func() {
		spinner := widget.NewActivity()
		spinner.Start()
		placeholder := container.NewHBox(
			widget.NewLabel(fmt.Sprintf("Collecting data for: %s [%d] (T=%v)...", name, pid, t)),
			spinner,
		)
		f.content.Add(placeholder)
		for {
			func() {
				stats, err := p.MemoryInfoEx()
				if err != nil {
					log.Fatal(err)
				}
				cpu, err := p.Percent(0)
				if err != nil {
					log.Fatal(err)
				}
				vv = append(vv, sample{
					memory:    int(stats.RSS - stats.Shared),
					cpu:       cpu / float64(f.numCores),
					timestamp: time.Now(),
				})
				if len(vv) == 0 {
					return
				}
				title := fmt.Sprintf("%s [%d] (T=%v)", name, pid, t)
				c, err := f.makeChart(title, vv)
				if err != nil {
					return
				}
				spinner.Stop()
				f.content.RemoveAll()
				f.content.Add(c)
			}()
			f.content.Refresh()
			select {
			case <-ticker.C:
			case <-closeC:
				log.Println("watcher closed")
				return
			}
		}
	}()
	return nil
}

type sample struct {
	memory    int
	cpu       float64
	timestamp time.Time
}

func (f *ChartFrame) makeChart(title string, d []sample) (fyne.CanvasObject, error) {
	s := f.u.w.Canvas().Scale()
	w, h := float32(900), float32(450)
	content, err := makeRawChart(d, int(s*w), int(s*h))
	if err != nil {
		return nil, err
	}
	r := fyne.NewStaticResource("Generated chart", content)
	chart := canvas.NewImageFromResource(r)
	chart.FillMode = canvas.ImageFillContain
	chart.SetMinSize(fyne.Size{Width: w, Height: h})
	t := widget.NewRichTextFromMarkdown("## " + title)
	return container.NewBorder(
		container.NewHBox(layout.NewSpacer(), t, layout.NewSpacer()),
		nil,
		nil,
		nil,
		chart,
	), nil
}

func makeRawChart(d []sample, w, h int) ([]byte, error) {
	xTime := make([]time.Time, len(d))
	yMemory := make([]float64, len(d))
	yCPU := make([]float64, len(d))
	for i, v := range d {
		xTime[i] = v.timestamp
		yMemory[i] = float64(v.memory)
		yCPU[i] = math.Round(v.cpu)
	}
	series := []chart.Series{
		chart.TimeSeries{
			Name:    "Memory",
			Style:   chart.StyleTextDefaults(),
			XValues: xTime,
			YValues: yMemory,
			YAxis:   chart.YAxisPrimary,
		},
		chart.TimeSeries{
			Name:    "CPU",
			Style:   chart.StyleTextDefaults(),
			XValues: xTime,
			YValues: yCPU,
			YAxis:   chart.YAxisSecondary,
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
			},
		},
		YAxisSecondary: chart.YAxis{
			Name:  "CPU%",
			Style: defaultStyle,
			ValueFormatter: func(x any) string {
				v := x.(float64)
				return strconv.Itoa(int(v))
			},
		},
		Series: series,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph, defaultStyle),
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
