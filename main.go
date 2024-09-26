package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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

type sample struct {
	memory    int
	timestamp time.Time
}

func main() {
	intervalFlag := flag.Int("t", 3, "Length of sampling interval in seconds")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Need to provide PID")
	}
	pid, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatal(err)
	}
	name, err := p.Name()
	if err != nil {
		log.Fatal(err)
	}

	a := app.New()
	w := a.NewWindow("Process monitor")
	vv := make([]sample, 0)
	page := container.NewVBox()
	t := time.Duration(*intervalFlag) * time.Second
	ticker := time.NewTicker(t)
	go func() {
		for {
			stats, err := p.MemoryInfoEx()
			if err != nil {
				log.Fatal(err)
			}
			vv = append(vv, sample{
				memory:    int(stats.RSS - stats.Shared),
				timestamp: time.Now(),
			})
			page.RemoveAll()
			if len(vv) > 1 {
				c, err := makeChart(name, vv, t)
				if err != nil {
					page.Add(widget.NewLabel("Not enough data points yet"))
				} else {
					page.Add(c)
				}
			} else {
				page.Add(widget.NewLabel("Not enough data points yet"))
			}
			page.Refresh()
			<-ticker.C
		}
	}()
	w.SetContent(page)
	w.ShowAndRun()
}

func makeChart(name string, d []sample, t time.Duration) (fyne.CanvasObject, error) {
	content, err := makeRawChart(name, d)
	if err != nil {
		return nil, err
	}
	r := fyne.NewStaticResource("dummy", content)
	chart := canvas.NewImageFromResource(r)
	chart.FillMode = canvas.ImageFillOriginal
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

func makeRawChart(name string, d []sample) ([]byte, error) {
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
		Width:  1200,
		Height: 600,
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
