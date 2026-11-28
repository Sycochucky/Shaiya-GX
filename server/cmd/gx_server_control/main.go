// Command gx_server_control is the standalone local GX Server Control desktop app.
package main

import (
	"context"
	"fmt"
	"log"
	"maps"
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
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
)

type serviceRow struct {
	status, endpoint     *widget.Label
	dot                  *canvas.Circle
	start, stop, restart *widget.Button
	force                *widget.Button
}

type dashboard struct {
	window                   fyne.Window
	manager                  *control.Manager
	controllerRoot           string
	ctx                      context.Context
	rows                     map[control.Service]serviceRow
	status                   map[control.Service]control.Status
	overall, players, notice *widget.Label
	shieldStatus             *widget.Label
	gameStatus               *widget.Label
	ownerNotice              *widget.Label
	setupIssue               error
	availableOwners          []string
	openInstallation         *widget.Button
	ownerPanel               *fyne.Container
	connectOwner             *widget.Button
	runningOwner             *control.Manager
	useInstallation          func(*control.Manager)
	overallDot               *canvas.Circle
	start, stop, restart     *widget.Button
	game, chooseGame         *widget.Button
	servers, shieldButton    *widget.Button
	tabs                     *container.AppTabs
	accountsItem             *container.TabItem
	selectedGame             *widget.Label
	busy                     bool
	gameBusy                 bool
	checked                  bool
	revision                 atomic.Uint64
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
	mountDashboard(window, ctx, rememberedInstallation(manager), manager.Root())
	window.ShowAndRun()
}

func mountDashboard(window fyne.Window, parent context.Context, manager *control.Manager, controllerRoot string) {
	ctx, cancel := context.WithCancel(parent)
	d := &dashboard{window: window, manager: manager, controllerRoot: controllerRoot, ctx: ctx, rows: make(map[control.Service]serviceRow),
		status: make(map[control.Service]control.Status)}
	d.useInstallation = func(owner *control.Manager) {
		rememberErr := rememberInstallation(d.controllerRoot, owner.Root())
		cancel()
		mountDashboard(window, parent, owner, d.controllerRoot)
		if rememberErr != nil {
			dialog.ShowInformation("GX installation", "Connected for this session. Could not save the selected folder: "+rememberErr.Error(), window)
		}
	}
	window.SetCloseIntercept(func() {
		d.closeDashboard(func() {
			cancel()
			window.Close()
		})
	})
	d.build()
	go d.poll()
	d.discoverRunningOwner()
}

func (d *dashboard) build() {
	d.setupIssue = d.manager.CheckInstallation()
	root := widget.NewLabel(d.manager.Root())
	root.Truncation = fyne.TextTruncateEllipsis
	d.overall = strong("Checking services")
	d.overallDot = canvas.NewCircle(quiet)
	d.players = widget.NewLabel("Players online: —")
	d.servers = widget.NewButton("Switch installation…", d.chooseInstallation)
	d.game = widget.NewButtonWithIcon("Run Game", theme.MediaPlayIcon(), d.runGame)
	d.game.Importance = widget.HighImportance
	d.chooseGame = widget.NewButton("Choose Game…", func() { d.pickGame(nil) })
	d.gameStatus = widget.NewLabel("")
	d.gameStatus.Wrapping = fyne.TextWrapWord
	d.gameStatus.Hide()
	heading := canvas.NewText("GX Server Control", ink)
	heading.TextSize = 22
	heading.TextStyle = fyne.TextStyle{Bold: true}
	status := container.NewHBox(container.NewCenter(statusCell(d.overallDot, d.overall)), container.NewCenter(d.players))
	identity := container.NewVBox(container.NewBorder(nil, nil, nil, status, heading), root)
	controls := container.NewBorder(identity, container.NewVBox(d.gameStatus, container.NewHBox(d.game)), nil, nil)
	installation := container.NewBorder(nil, nil, newBanner(d.ctx, d.manager.Root()), nil, container.NewPadded(controls))

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
			force:   widget.NewButton("Force stop", func() { d.run(control.ForceStop, []control.Service{service}) }),
		}
		row.start.Importance = widget.HighImportance
		row.endpoint.TextStyle.Monospace = true
		row.endpoint.Truncation = fyne.TextTruncateEllipsis
		d.rows[service] = row
		d.status[service] = control.Status{State: control.Unknown}
		serviceContents = append(serviceContents, widget.NewSeparator(),
			container.NewGridWithColumns(4, strong(service.Label()), statusCell(row.dot, row.status), row.endpoint,
				container.NewHBox(row.start, row.stop, row.restart, row.force)))
	}
	d.ownerNotice = widget.NewLabel("")
	d.ownerNotice.Wrapping = fyne.TextWrapWord
	d.connectOwner = widget.NewButton("Reconnect", func() {
		if d.busy || d.useInstallation == nil {
			return
		}
		if d.runningOwner != nil {
			d.useInstallation(d.runningOwner)
			return
		}
		d.chooseInstallation()
	})
	d.openInstallation = widget.NewButton("Open Installation…", d.chooseInstallation)
	d.ownerPanel = container.NewBorder(nil, nil, nil,
		container.NewHBox(d.connectOwner, d.openInstallation), d.ownerNotice)
	d.openInstallation.Hide()
	d.connectOwner.Hide()
	d.ownerPanel.Hide()
	if d.setupIssue != nil {
		d.openInstallation.Show()
		d.ownerNotice.SetText("Open a complete GX Server folder to manage services.")
		d.ownerPanel.Show()
	}
	d.shieldStatus = widget.NewLabel("Checking application shield")
	d.shieldStatus.Wrapping = fyne.TextWrapWord
	d.shieldButton = widget.NewButton("Enable Under Attack for 30 min", func() { d.applyShield(30) })
	d.shieldButton.Importance = widget.HighImportance
	shield := panel(container.NewVBox(
		container.NewBorder(nil, nil, strong("GX Shield"), d.shieldButton, nil),
		d.shieldStatus,
	))
	d.notice = widget.NewLabel("")
	d.notice.Wrapping = fyne.TextWrapWord
	d.notice.Hide()
	serviceContents = append(serviceContents, d.notice)
	services := panel(container.NewVBox(serviceContents...))
	overview := container.NewTabItem("Overview", container.NewVScroll(container.NewVBox(services, shield)))
	activity := container.NewTabItem("Activity", container.NewStack())
	d.accountsItem = container.NewTabItem("Accounts", container.NewStack())
	settings := container.NewTabItem("Settings", container.NewStack())
	setup := container.NewTabItem("Setup", container.NewStack())
	d.tabs = container.NewAppTabs(overview, activity, d.accountsItem, settings, setup)
	pages := map[*container.TabItem]func() fyne.CanvasObject{
		activity:       func() fyne.CanvasObject { return newLogView(d.window, d.ctx, d.manager.Root()).content },
		d.accountsItem: d.accountView,
		settings:       d.settingsView,
		setup:          d.setupView,
	}
	d.tabs.OnSelected = func(item *container.TabItem) {
		if build := pages[item]; build != nil {
			delete(pages, item)
			item.Content = build()
			d.tabs.Refresh()
		}
	}
	d.window.SetContent(container.NewPadded(container.NewBorder(
		container.NewVBox(panel(installation), d.ownerPanel), nil, nil, nil, d.tabs)))
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
				if d.checked && maps.Equal(d.status, status) {
					return
				}
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

func (d *dashboard) discoverRunningOwner() {
	if d.useInstallation == nil {
		return
	}
	go func() {
		if d.setupIssue != nil {
			owners, err := d.manager.RunningInstallations(d.ctx)
			if err == nil && len(owners) > 0 {
				fyne.Do(func() {
					if d.ctx.Err() != nil {
						return
					}
					d.availableOwners = owners
					d.ownerNotice.SetText("GX is already running. Reconnect to a verified installation.")
					d.connectOwner.Show()
					d.ownerPanel.Refresh()
				})
			}
			return
		}
		status := d.manager.Snapshot(d.ctx)
		if status[control.Login].Processes+status[control.World].Processes != 0 || d.ctx.Err() != nil {
			return
		}
		owner, err := d.manager.ConflictingInstallation(d.ctx)
		if err != nil {
			return
		}
		fyne.Do(func() {
			if d.ctx.Err() != nil || d.busy ||
				d.status[control.Login].Processes+d.status[control.World].Processes != 0 {
				return
			}
			d.runningOwner = owner
			if owner == nil {
				d.ownerPanel.Hide()
				return
			}
			d.ownerNotice.SetText("GX is already running from " + owner.Root())
			d.connectOwner.Show()
			d.ownerPanel.Show()
		})
	}()
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
		if d.setupIssue != nil && info.Processes == 0 {
			row.status.SetText("Unavailable")
			endpoint = ""
		}
		if endpoint == "" {
			endpoint = "—"
		}
		row.endpoint.SetText(endpoint)
		canStart := d.setupIssue == nil && (info.State == control.Stopped || info.State == control.Failed)
		canStop := info.Processes != 0
		if canStop {
			row.start.Hide()
			row.stop.Show()
			row.restart.Show()
			row.force.Show()
		} else {
			if d.setupIssue == nil {
				row.start.Show()
			} else {
				row.start.Hide()
			}
			row.stop.Hide()
			row.restart.Hide()
			row.force.Hide()
		}
		setEnabled(row.start, !d.busy && canStart)
		setEnabled(row.stop, !d.busy && canStop)
		setEnabled(row.restart, !d.busy && canStop && d.setupIssue == nil)
		setEnabled(row.force, !d.busy && canStop)
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
	case d.setupIssue != nil:
		d.overall.SetText("Setup required")
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
	if d.setupIssue != nil {
		detail = d.setupIssue.Error()
	}
	if detail != "" {
		d.notice.SetText(detail)
		d.notice.Show()
	} else {
		d.notice.Hide()
	}
	setEnabled(d.start, !d.busy && possibleStart)
	setEnabled(d.stop, !d.busy && possibleStop)
	setEnabled(d.restart, !d.busy && possibleStop && d.setupIssue == nil)
	setEnabled(d.servers, !d.busy)
	setEnabled(d.game, !d.busy && !d.gameBusy && d.setupIssue == nil)
	setEnabled(d.chooseGame, !d.busy && d.setupIssue == nil)
	setEnabled(d.openInstallation, !d.busy)
	setEnabled(d.connectOwner, !d.busy)
	d.refreshShield()
}

func (d *dashboard) refreshShield() {
	login, world := d.status[control.Login], d.status[control.World]
	if login.State != control.Ready || world.State != control.Ready || login.Shield.Mode == "" || world.Shield.Mode == "" {
		d.shieldStatus.SetText("Shield status is available when both services are ready.")
		d.shieldButton.Disable()
		return
	}
	if login.Shield.Mode != world.Shield.Mode {
		d.shieldStatus.SetText("Login and World shield modes differ. Apply Under Attack to synchronize them.")
		d.shieldButton.SetText("Apply Under Attack to both")
		d.shieldButton.OnTapped = func() { d.applyShield(30) }
		setEnabled(d.shieldButton, !d.busy)
		return
	}
	if login.Shield.Mode == netx.ShieldUnderAttack {
		until := login.Shield.Until
		if world.Shield.Until.Before(until) {
			until = world.Shield.Until
		}
		trigger := "Manual"
		if login.Shield.Automatic || world.Shield.Automatic {
			trigger = "Automatic"
		}
		d.shieldStatus.SetText(fmt.Sprintf("%s Under Attack mode until %s. Rejects: Login %d, World %d. Existing players remain connected.", trigger, until.Local().Format("15:04:05"), login.Shield.RejectedTotal, world.Shield.RejectedTotal))
		d.shieldButton.SetText("Return to standard limits")
		d.shieldButton.OnTapped = func() {
			dialog.ShowConfirm("Return to standard limits", "Remove the temporary stricter admission limits from both services?", func(ok bool) {
				if ok {
					d.applyShield(0)
				}
			}, d.window)
		}
	} else {
		d.shieldStatus.SetText("Standard admission limits are active on both services.")
		d.shieldButton.SetText("Enable Under Attack for 30 min")
		d.shieldButton.OnTapped = func() { d.applyShield(30) }
	}
	setEnabled(d.shieldButton, !d.busy)
}

func (d *dashboard) applyShield(minutes int) {
	if d.busy {
		return
	}
	d.busy = true
	d.revision.Add(1)
	d.refresh()
	go func() {
		err := d.manager.SetShield(d.ctx, minutes)
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

func (d *dashboard) closeDashboard(closeWindow func()) {
	if d.busy {
		dialog.ShowInformation("Operation in progress", "Wait for the service operation to finish.", d.window)
		return
	}
	if d.status[control.Login].Processes+d.status[control.World].Processes == 0 {
		closeWindow()
		return
	}
	keep := widget.NewButton("Keep services running", nil)
	stop := widget.NewButton("Stop and exit", nil)
	confirm := dialog.NewCustom("Close GX Server Control", "Cancel", container.NewVBox(
		widget.NewLabel("Reopening the dashboard reconnects to services left running."),
		widget.NewLabel("Stopping World disconnects players."), container.NewHBox(keep, stop)), d.window)
	keep.OnTapped = func() { confirm.Hide(); closeWindow() }
	stop.OnTapped = func() {
		confirm.Hide()
		d.applyOperation(control.Stop, control.Services(), closeWindow)
	}
	confirm.Show()
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
	if action == control.ForceStop {
		message += "\n\nForce stop ends the process immediately. Unsaved progress can be lost. Use Stop for a normal shutdown."
	}
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
	d.applyOperation(action, targets, nil)
}

func (d *dashboard) applyOperation(action control.Action, targets []control.Service, onSuccess func()) {
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
				if action == control.Start {
					d.discoverRunningOwner()
				}
			} else if onSuccess != nil {
				onSuccess()
			}
		})
	}()
}
