package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func (d *dashboard) ipBanView() fyne.CanvasObject {
	p := &accountPanel{d: d, ctx: d.ctx, status: widget.NewLabel("")}
	query := widget.NewEntry()
	query.SetPlaceHolder("Search banned IP addresses")
	details := container.NewVBox(widget.NewLabel("Select an address."))
	var rows []operatorwire.IPBan
	list := widget.NewList(func() int { return len(rows) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(rows[i].Address)
		})
	search := widget.NewButton("Search", nil)
	add := widget.NewButton("Ban IP…", nil)
	refresh := func() {
		if p.busy {
			return
		}
		query.Disable()
		search.Disable()
		add.Disable()
		p.request(operatorwire.Request{Action: "ip-bans", Query: strings.TrimSpace(query.Text)}, func(res operatorwire.Response, err error) {
			query.Enable()
			search.Enable()
			add.Enable()
			if err != nil {
				return
			}
			list.UnselectAll()
			rows = res.IPBans
			list.Refresh()
			details.Objects = []fyne.CanvasObject{widget.NewLabel("Select an address.")}
			details.Refresh()
			switch {
			case res.More:
				p.status.SetText("Showing the first 100 matches. Narrow the search to find more.")
			case len(rows) == 0:
				p.status.SetText("No matching IP bans.")
			default:
				p.status.SetText(fmt.Sprintf("%d banned addresses", len(rows)))
			}
		})
	}
	list.OnSelected = func(i widget.ListItemID) {
		if i < 0 || i >= len(rows) {
			return
		}
		ban := rows[i]
		reason := widget.NewLabel(ban.Reason)
		reason.Wrapping = fyne.TextWrapWord
		remove := widget.NewButton("Unban IP", func() {
			confirm := dialog.NewConfirm("Unban IP address", "Remove the ban for "+ban.Address+"?", func(ok bool) {
				if ok {
					p.request(operatorwire.Request{Action: "unban-ip", IPAddress: ban.Address}, func(_ operatorwire.Response, err error) {
						if err == nil {
							refresh()
						}
					})
				}
			}, d.window)
			confirm.SetConfirmText("Unban")
			confirm.Show()
		})
		details.Objects = []fyne.CanvasObject{strong(ban.Address),
			widget.NewLabel("Banned " + ban.CreatedAt.Local().Format("2 Jan 2006 at 15:04")), reason, container.NewHBox(remove)}
		details.Refresh()
	}
	add.OnTapped = func() {
		address, reason := widget.NewEntry(), widget.NewEntry()
		address.SetPlaceHolder("IPv4 or IPv6 address")
		p.form("Ban IP address", "Ban IP", []*widget.FormItem{
			widget.NewFormItem("IP address", address), widget.NewFormItem("Reason (optional)", reason),
		}, "This blocks Login and World and disconnects players on this address. Players may share an IP.",
			func() (operatorwire.Request, error) {
				return operatorwire.Request{Action: "ban-ip", IPAddress: address.Text, BanReason: reason.Text}, nil
			}, func() {
				query.SetText("")
				refresh()
			})
	}
	search.OnTapped = refresh
	query.OnSubmitted = func(string) { refresh() }
	toolbar := container.NewBorder(nil, nil, nil, container.NewHBox(search, add), query)
	split := container.NewHSplit(panel(list), panel(container.NewVScroll(details)))
	split.SetOffset(0.36)
	refresh()
	return container.NewBorder(toolbar, p.status, nil, nil, split)
}
