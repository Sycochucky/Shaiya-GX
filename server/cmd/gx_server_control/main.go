// Command gx_server_control is the standalone local GX Server Control desktop app.
package main

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

type serviceRow struct {
	status, endpoint     *widget.Label
	dot                  *canvas.Circle
	start, stop, restart *widget.Button
}

type dashboard struct {
	window                         fyne.Window
	manager                        *control.Manager
	ctx                            context.Context
	rows                           map[control.Service]serviceRow
	status                         map[control.Service]control.Status
	overall, players, notice       *widget.Label
	overallDot                     *canvas.Circle
	start, stop, restart, settings *widget.Button
	busy                           bool
	checked                        bool
	revision                       atomic.Uint64
}

func main() {
	manager, err := control.Installation()
	if err != nil {
		log.Fatal(err)
	}
	instance := app.NewWithID("com.shaiyagx.server-control")
	instance.Settings().SetTheme(controlTheme{Theme: theme.DefaultTheme()})
	window := instance.NewWindow("GX Server Control")
	window.Resize(fyne.NewSize(1120, 760))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := &dashboard{window: window, manager: manager, ctx: ctx, rows: make(map[control.Service]serviceRow),
		status: make(map[control.Service]control.Status)}
	window.SetCloseIntercept(func() {
		if d.busy {
			dialog.ShowInformation("Operation in progress", "Wait for the service operation to finish.", window)
			return
		}
		cancel()
		window.Close()
	})
	d.build()
	go d.poll()
	window.ShowAndRun()
}

func (d *dashboard) build() {
	root := widget.NewLabel(d.manager.Root())
	root.Truncation = fyne.TextTruncateEllipsis
	d.overall = strong("Checking services")
	d.overallDot = canvas.NewCircle(quiet)
	d.players = widget.NewLabel("Players online: —")
	d.settings = widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), d.openSettings)
	installation := container.NewBorder(nil, nil, nil,
		container.NewHBox(container.NewCenter(statusCell(d.overallDot, d.overall)),
			container.NewCenter(d.players), container.NewCenter(d.settings)),
		container.NewVBox(strong("Local installation"), root))

	d.start = widget.NewButton("Start All", func() { d.run(control.Start, control.Services()) })
	d.stop = widget.NewButton("Stop All", func() { d.run(control.Stop, control.Services()) })
	d.restart = widget.NewButton("Restart All", func() { d.run(control.Restart, control.Services()) })
	serviceTitle := container.NewBorder(nil, nil, strong("Services"),
		container.NewHBox(d.start, d.stop, d.restart), nil)
	serviceContents := []fyne.CanvasObject{
		serviceTitle, widget.NewSeparator(),
		container.NewGridWithColumns(4, strong("Service"), strong("Status"), strong("Endpoint"), strong("Actions")),
	}
	for _, service := range control.Services() {
		row := serviceRow{
			status: strong(string(control.Unknown)), endpoint: widget.NewLabel("—"), dot: canvas.NewCircle(quiet),
			start:   widget.NewButton("Start", func() { d.run(control.Start, []control.Service{service}) }),
			stop:    widget.NewButton("Stop", func() { d.run(control.Stop, []control.Service{service}) }),
			restart: widget.NewButton("Restart", func() { d.run(control.Restart, []control.Service{service}) }),
		}
		row.start.Importance = widget.HighImportance
		row.endpoint.TextStyle.Monospace = true
		row.endpoint.Truncation = fyne.TextTruncateEllipsis
		d.rows[service] = row
		d.status[service] = control.Status{State: control.Unknown}
		serviceContents = append(serviceContents, widget.NewSeparator(),
			container.NewGridWithColumns(4, strong(service.Label()), statusCell(row.dot, row.status), row.endpoint,
				container.NewHBox(row.start, row.stop, row.restart)))
	}
	d.notice = widget.NewLabel("")
	d.notice.Wrapping = fyne.TextWrapWord
	d.notice.Hide()
	serviceContents = append(serviceContents, d.notice)
	services := panel(container.NewVBox(serviceContents...))
	logs := newLogView(d.window, d.ctx, d.manager.Root())
	d.window.SetContent(container.NewPadded(container.NewBorder(
		container.NewVBox(panel(installation), services), nil, nil, nil, logs.content)))
	d.refresh()
}

func strong(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func statusCell(dot *canvas.Circle, label *widget.Label) fyne.CanvasObject {
	return container.NewHBox(container.NewCenter(container.NewGridWrap(fyne.NewSize(9, 9), dot)), label)
}

func (d *dashboard) poll() {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for d.ctx.Err() == nil {
		revision := d.revision.Load()
		status := d.manager.Snapshot(d.ctx)
		if d.ctx.Err() != nil {
			return
		}
		fyne.Do(func() {
			if d.ctx.Err() == nil && !d.busy && d.revision.Load() == revision {
				d.status = status
				d.checked = true
				d.refresh()
			}
		})
		select {
		case <-d.ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func (d *dashboard) refresh() {
	ready, possibleStart, possibleStop, failed := 0, false, false, false
	detail := ""
	for _, service := range control.Services() {
		info, row := d.status[service], d.rows[service]
		row.dot.FillColor = statusColor(info.State)
		row.dot.Refresh()
		row.status.SetText(string(info.State))
		endpoint := info.Endpoint
		if endpoint == "" {
			endpoint = "—"
		}
		row.endpoint.SetText(endpoint)
		canStart := info.State == control.Stopped || info.State == control.Failed
		canStop := info.PID != 0
		if canStop {
			row.start.Hide()
			row.stop.Show()
			row.restart.Show()
		} else {
			row.start.Show()
			row.stop.Hide()
			row.restart.Hide()
		}
		setEnabled(row.start, !d.busy && canStart)
		setEnabled(row.stop, !d.busy && canStop)
		setEnabled(row.restart, !d.busy && canStop)
		possibleStart = possibleStart || canStart
		possibleStop = possibleStop || canStop
		if info.State == control.Ready {
			ready++
		}
		if info.State == control.Failed || info.State == control.Unknown {
			failed = true
		}
		if detail == "" && info.Detail != "" {
			detail = service.Label() + ": " + info.Detail
		}
	}
	switch {
	case d.busy:
		d.overall.SetText("Operation in progress")
		d.overallDot.FillColor = warning
	case !d.checked:
		d.overall.SetText("Checking services")
		d.overallDot.FillColor = quiet
	case ready == 2:
		d.overall.SetText("Operational")
		d.overallDot.FillColor = healthy
	case ready == 1 || failed:
		d.overall.SetText("Degraded")
		d.overallDot.FillColor = warning
	default:
		d.overall.SetText("Stopped")
		d.overallDot.FillColor = quiet
	}
	d.overallDot.Refresh()
	if world := d.status[control.World]; world.State == control.Ready {
		d.players.SetText(fmt.Sprintf("Players online: %d", world.Players))
	} else {
		d.players.SetText("Players online: —")
	}
	if detail != "" {
		d.notice.SetText(detail)
		d.notice.Show()
	} else {
		d.notice.Hide()
	}
	setEnabled(d.start, !d.busy && possibleStart)
	setEnabled(d.stop, !d.busy && possibleStop)
	setEnabled(d.restart, !d.busy && possibleStop)
	setEnabled(d.settings, !d.busy)
}

func setEnabled(button *widget.Button, enabled bool) {
	if enabled {
		button.Enable()
	} else {
		button.Disable()
	}
}

func (d *dashboard) run(action control.Action, targets []control.Service) {
	if d.busy {
		return
	}
	target := "all services"
	if len(targets) == 1 {
		target = targets[0].Label()
	}
	if action == control.Start {
		d.apply(action, targets)
		return
	}
	message := fmt.Sprintf("%s %s for this installation?", action, target)
	for _, service := range targets {
		if service == control.World {
			message += "\n\nStopping World disconnects any players online. "
			if world := d.status[service]; world.State == control.Ready {
				message += fmt.Sprintf("%d currently online.", world.Players)
			} else {
				message += "The player count is unavailable."
			}
		}
	}
	dialog.ShowConfirm(string(action)+" "+target, message, func(ok bool) {
		if ok {
			d.apply(action, targets)
		}
	}, d.window)
}

func (d *dashboard) apply(action control.Action, targets []control.Service) {
	d.busy = true
	d.revision.Add(1)
	d.refresh()
	go func() {
		err := d.manager.Apply(d.ctx, action, targets, func(service control.Service, info control.Status) {
			fyne.Do(func() {
				if info.State == control.Starting || info.State == control.Stopping {
					info.Endpoint = d.status[service].Endpoint
				}
				d.status[service] = info
				d.refresh()
			})
		})
		status := d.manager.Snapshot(d.ctx)
		fyne.Do(func() {
			d.revision.Add(1)
			d.busy = false
			d.status = status
			d.refresh()
			if err != nil {
				dialog.ShowError(err, d.window)
			}
		})
	}()
}
