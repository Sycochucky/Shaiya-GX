package main

import (
	"context"
	"slices"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

const logHistory = 1000

type logView struct {
	content                   fyne.CanvasObject
	list                      *widget.List
	footer, empty             *widget.Label
	service, level, readError string
	all, visible              []control.LogEntry
	paused                    bool
}

func newLogView(window fyne.Window, ctx context.Context, root string) *logView {
	v := &logView{service: "All", level: "All levels", footer: widget.NewLabel("Following live")}
	v.list = widget.NewList(func() int { return len(v.visible) }, newLogRow, v.renderRow)
	v.list.OnSelected = func(index widget.ListItemID) {
		v.list.UnselectAll()
		if index < 0 || index >= len(v.visible) {
			return
		}
		entry := v.visible[index]
		message := widget.NewLabel(entry.Message)
		message.Wrapping = fyne.TextWrapWord
		details := dialog.NewCustom(entry.Service.Label()+" · "+strings.ToUpper(entry.Level), "Close",
			container.NewVScroll(message), window)
		details.Resize(fyne.NewSize(620, 250))
		details.Show()
	}
	filter := widget.NewRadioGroup([]string{"All", "Login", "World"}, func(service string) {
		v.service = service
		v.refresh()
	})
	filter.Horizontal = true
	filter.Required = true
	filter.SetSelected("All")
	level := widget.NewSelect([]string{"All levels", "Error", "Warning", "Info", "Debug"}, func(level string) {
		v.level = level
		v.refresh()
	})
	level.SetSelected("All levels")
	pause := widget.NewButton("Pause follow", nil)
	pause.OnTapped = func() {
		v.paused = !v.paused
		if v.paused {
			pause.SetText("Resume follow")
		} else {
			pause.SetText("Pause follow")
		}
		v.refresh()
	}
	header := container.New(&logColumns{}, strong("Time"), strong("Level"), strong("Service"), strong("Message"))
	toolbar := container.NewBorder(nil, nil, filter, container.NewHBox(level, pause), nil)
	v.empty = widget.NewLabel("Waiting for service logs.")
	v.content = panel(container.NewBorder(
		container.NewVBox(strong("Live Logs"), toolbar, widget.NewSeparator(), header, widget.NewSeparator()),
		container.NewBorder(widget.NewSeparator(), nil, nil, v.footer, nil), nil, nil,
		container.NewStack(v.list, container.NewCenter(v.empty))))
	v.refresh()
	go v.poll(ctx, control.NewLogReader(root))
	return v
}

func (v *logView) poll(ctx context.Context, reader *control.LogReader) {
	tick := time.NewTicker(400 * time.Millisecond)
	defer tick.Stop()
	for ctx.Err() == nil {
		records, err := reader.Poll()
		readError := ""
		if err != nil {
			readError = err.Error()
		}
		if ctx.Err() != nil {
			return
		}
		fyne.Do(func() {
			if ctx.Err() != nil {
				return
			}
			if len(records) == 0 && readError == v.readError {
				v.follow()
				return
			}
			v.all = append(v.all, records...)
			slices.SortStableFunc(v.all, func(first, second control.LogEntry) int { return first.Time.Compare(second.Time) })
			if extra := len(v.all) - logHistory; extra > 0 {
				v.all = slices.Delete(v.all, 0, extra)
			}
			v.readError = readError
			v.refresh()
		})
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func (v *logView) refresh() {
	v.visible = v.visible[:0]
	level := strings.ToLower(v.level)
	if level == "warning" {
		level = "warn"
	}
	for _, record := range v.all {
		if (v.service == "All" || record.Service.Label() == v.service) &&
			(v.level == "All levels" || record.Level == level) {
			v.visible = append(v.visible, record)
		}
	}
	v.list.Refresh()
	if v.empty != nil {
		if len(v.visible) == 0 {
			if len(v.all) > 0 {
				v.empty.SetText("No logs match these filters.")
			} else {
				v.empty.SetText("Waiting for service logs.")
			}
			v.empty.Show()
		} else {
			v.empty.Hide()
		}
	}
	switch {
	case v.readError != "":
		v.footer.SetText(v.readError)
	case v.paused:
		v.footer.SetText("Follow paused")
	default:
		v.footer.SetText("Following live")
	}
	v.follow()
}

func (v *logView) follow() {
	if !v.paused && len(v.visible) > 0 {
		v.list.ScrollToBottom()
	}
}

func newLogRow() fyne.CanvasObject {
	labels := make([]fyne.CanvasObject, 4)
	for i := range labels {
		label := widget.NewLabel("")
		label.TextStyle.Monospace = true
		label.Truncation = fyne.TextTruncateEllipsis
		labels[i] = label
	}
	return container.NewStack(canvas.NewRectangle(surface), container.New(&logColumns{}, labels...))
}

func (v *logView) renderRow(id widget.ListItemID, object fyne.CanvasObject) {
	if id < 0 || id >= len(v.visible) {
		return
	}
	entry := v.visible[id]
	row := object.(*fyne.Container)
	background := row.Objects[0].(*canvas.Rectangle)
	background.FillColor = surface
	if entry.Level == "error" {
		background.FillColor = faultSurface
	}
	background.Refresh()
	stamp := "—"
	if !entry.Time.IsZero() {
		stamp = entry.Time.Local().Format("15:04:05")
	}
	labels := row.Objects[1].(*fyne.Container).Objects
	for i, text := range []string{stamp, strings.ToUpper(entry.Level), entry.Service.Label(), strings.ReplaceAll(entry.Message, "\n", " ")} {
		labels[i].(*widget.Label).SetText(text)
	}
}

type logColumns struct{}

var logWidths = [...]float32{95, 80, 95}

func (*logColumns) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	var offset float32
	for i, width := range logWidths {
		objects[i].Move(fyne.NewPos(offset, 0))
		objects[i].Resize(fyne.NewSize(width, size.Height))
		offset += width
	}
	objects[3].Move(fyne.NewPos(offset, 0))
	objects[3].Resize(fyne.NewSize(max(0, size.Width-offset), size.Height))
}

func (*logColumns) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width, height float32
	for _, column := range logWidths {
		width += column
	}
	for _, item := range objects {
		height = max(height, item.MinSize().Height)
	}
	return fyne.NewSize(width+160, height)
}
