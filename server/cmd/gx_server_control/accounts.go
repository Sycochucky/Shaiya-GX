package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/internal/pvprank"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

type accountPanel struct {
	d             *dashboard
	ctx           context.Context
	query         *widget.Entry
	status        *widget.Label
	list          *widget.List
	details       *fyne.Container
	accounts      []operatorwire.Account
	busy          bool
	pendingSearch *string
}

func (d *dashboard) accountView() fyne.CanvasObject {
	p := &accountPanel{d: d, ctx: d.ctx, query: widget.NewEntry(), status: widget.NewLabel(""),
		details: container.NewVBox(widget.NewLabel("Select an account."))}
	p.query.SetPlaceHolder("Search accounts by username")
	p.status.Wrapping = fyne.TextWrapWord
	p.list = widget.NewList(func() int { return len(p.accounts) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, object fyne.CanvasObject) {
			a := p.accounts[i]
			label := a.Name
			if a.Banned {
				label += "   ·   Banned"
			}
			object.(*widget.Label).SetText(label)
		})
	p.list.OnSelected = func(i widget.ListItemID) {
		if p.busy || i < 0 || i >= len(p.accounts) {
			p.list.UnselectAll()
			return
		}
		p.show(p.accounts[i].ID)
	}
	search := widget.NewButton("Search", p.search)
	p.query.OnSubmitted = func(string) { p.search() }
	toolbar := container.NewBorder(nil, nil, nil, container.NewHBox(search, widget.NewButton("New account", p.create)), p.query)
	split := container.NewHSplit(panel(p.list), panel(container.NewVScroll(p.details)))
	split.SetOffset(0.36)
	players := container.NewTabItem("Players and characters", container.NewBorder(toolbar, p.status, nil, nil, split))
	ipBans := container.NewTabItem("IP bans", container.NewStack())
	tabs := container.NewAppTabs(players, ipBans)
	tabs.OnSelected = func(item *container.TabItem) {
		if item == ipBans {
			tabs.OnSelected = nil
			item.Content = d.ipBanView()
			tabs.Refresh()
		}
	}
	p.search()
	return tabs
}

func (p *accountPanel) request(req operatorwire.Request, complete func(operatorwire.Response, error)) {
	if p.busy {
		complete(operatorwire.Response{}, errors.New("another account operation is still running"))
		return
	}
	p.busy = true
	switch req.Action {
	case "accounts":
		p.status.SetText("Loading accounts…")
	case "details":
		p.status.SetText("Loading account and characters…")
	case "ip-bans":
		p.status.SetText("Loading IP bans…")
	default:
		p.status.SetText("Saving…")
	}
	go func() {
		res, err := p.d.manager.Operator(p.ctx, req)
		if p.ctx.Err() != nil {
			return
		}
		fyne.Do(func() {
			if p.ctx.Err() != nil {
				return
			}
			p.busy = false
			if req.Action == "accounts" && p.pendingSearch != nil {
				p.runPendingSearch()
				return
			}
			if err != nil {
				p.status.SetText(err.Error())
			} else {
				p.status.SetText("Updated.")
			}
			complete(res, err)
			if !p.busy {
				p.runPendingSearch()
			}
		})
	}()
}

func (p *accountPanel) search() {
	query := strings.TrimSpace(p.query.Text)
	if p.busy {
		p.pendingSearch = &query
		p.status.SetText("Search queued…")
		return
	}
	p.searchAccounts(query)
}

func (p *accountPanel) runPendingSearch() {
	if p.pendingSearch == nil {
		return
	}
	query := *p.pendingSearch
	p.pendingSearch = nil
	p.searchAccounts(query)
}

func (p *accountPanel) searchAccounts(query string) {
	p.request(operatorwire.Request{Action: "accounts", Query: query}, func(res operatorwire.Response, err error) {
		if err != nil {
			return
		}
		p.list.UnselectAll()
		p.accounts = res.Accounts
		p.list.Refresh()
		p.details.Objects = []fyne.CanvasObject{widget.NewLabel("Select an account.")}
		p.details.Refresh()
		if res.More {
			p.status.SetText("Showing the first 100 matches. Narrow the search to find more.")
		}
		if len(res.Accounts) == 0 {
			p.status.SetText("No matching accounts.")
		}
	})
}

func (p *accountPanel) show(id uint32) {
	p.request(operatorwire.Request{Action: "details", AccountID: id}, func(res operatorwire.Response, err error) {
		if err != nil {
			return
		}
		a := *res.Account
		for i, row := range p.accounts {
			if row.ID == a.ID {
				p.accounts[i] = a
				p.list.Refresh()
				break
			}
		}
		role := "Player"
		if a.GM {
			role = "Game master"
		}
		access := "Ban account"
		if a.Banned {
			access = "Unban account"
		}
		objects := []fyne.CanvasObject{
			strong(a.Name), widget.NewLabel(fmt.Sprintf("Account #%d  ·  %s  ·  %d Item Mall points", a.ID, role, a.Points)),
			container.NewHBox(widget.NewButton("Change password", func() { p.password(a) }),
				widget.NewButton("Role and points", func() { p.editAccount(a) }),
				widget.NewButton(access, func() { p.changeBan(a) }),
				widget.NewButton("Refresh", func() { p.show(a.ID) })),
		}
		if a.Banned {
			label := "Account is banned."
			if a.BanReason != "" {
				label = "Banned: " + a.BanReason
			}
			reason := widget.NewLabel(label)
			reason.Wrapping = fyne.TextWrapWord
			objects = append(objects, reason)
		}
		objects = append(objects, widget.NewSeparator(), strong("Characters"))
		if len(res.Characters) == 0 {
			objects = append(objects, widget.NewLabel("This account has no characters."))
		}
		for _, ch := range res.Characters {
			title := fmt.Sprintf("%s  ·  Level %d", ch.Name, ch.Level)
			var actions fyne.CanvasObject
			if ch.Deleted {
				title += "  ·  Deleted"
				actions = widget.NewButton("Restore", func() {
					p.request(operatorwire.Request{Action: "restore", AccountID: a.ID, CharacterID: ch.ID}, func(_ operatorwire.Response, err error) {
						if err == nil {
							p.show(a.ID)
						}
					})
				})
			} else {
				remove := widget.NewButton("Delete", func() { p.deleteCharacter(a, ch) })
				if ch.Online {
					title += "  ·  Online"
					remove.Disable()
				}
				actions = container.NewHBox(widget.NewButton("Edit stats", func() { p.editStats(a, ch) }), remove)
			}
			objects = append(objects, container.NewBorder(nil, nil, nil, actions, widget.NewLabel(title)), widget.NewSeparator())
		}
		p.details.Objects = objects
		p.details.Refresh()
	})
}

func (p *accountPanel) form(title, action string, fields []*widget.FormItem, help string,
	read func() (operatorwire.Request, error), after func()) {
	message := widget.NewLabel(help)
	message.Wrapping = fyne.TextWrapWord
	button := widget.NewButton(action, nil)
	button.Importance = widget.HighImportance
	body := container.NewVBox(widget.NewForm(fields...), message, container.NewHBox(button))
	popup := dialog.NewCustom(title, "Cancel", container.NewVScroll(body), p.d.window)
	height := min(720, float32(180+len(fields)*52))
	popup.Resize(fyne.NewSize(580, height))
	button.OnTapped = func() {
		req, err := read()
		if err != nil {
			message.SetText(err.Error())
			return
		}
		button.Disable()
		p.request(req, func(_ operatorwire.Response, err error) {
			if err != nil {
				message.SetText(err.Error())
				button.Enable()
				return
			}
			popup.Hide()
			after()
		})
	}
	popup.Show()
}

func (p *accountPanel) changeBan(a operatorwire.Account) {
	refresh := func() { p.show(a.ID) }
	if a.Banned {
		confirm := dialog.NewConfirm("Unban account", "Remove the dashboard ban for "+a.Name+"?", func(ok bool) {
			if ok {
				p.request(operatorwire.Request{Action: "unban-account", AccountID: a.ID}, func(_ operatorwire.Response, err error) {
					if err == nil {
						refresh()
					}
				})
			}
		}, p.d.window)
		confirm.SetConfirmText("Unban")
		confirm.Show()
		return
	}
	reason := widget.NewEntry()
	p.form("Ban "+a.Name, "Ban account", []*widget.FormItem{widget.NewFormItem("Reason (optional)", reason)},
		"This signs out this account immediately. The ban stays active until you remove it.",
		func() (operatorwire.Request, error) {
			return operatorwire.Request{Action: "ban-account", AccountID: a.ID, BanReason: reason.Text}, nil
		}, refresh)
}

func (p *accountPanel) create() {
	name, password, confirm := widget.NewEntry(), widget.NewPasswordEntry(), widget.NewPasswordEntry()
	gm := widget.NewCheck("Game master", nil)
	p.form("Create account", "Create", []*widget.FormItem{
		widget.NewFormItem("Username", name), widget.NewFormItem("Password", password),
		widget.NewFormItem("Repeat password", confirm), widget.NewFormItem("Role", gm),
	}, "Usernames can use up to 18 UTF-8 bytes.", func() (operatorwire.Request, error) {
		if password.Text == "" || password.Text != confirm.Text {
			return operatorwire.Request{}, errors.New("enter matching passwords")
		}
		return operatorwire.Request{Action: "create", Name: strings.TrimSpace(name.Text), Password: password.Text, GM: gm.Checked}, nil
	}, func() { p.query.SetText(strings.TrimSpace(name.Text)); p.search() })
}

func (p *accountPanel) password(a operatorwire.Account) {
	password, confirm := widget.NewPasswordEntry(), widget.NewPasswordEntry()
	p.form("Change password for "+a.Name, "Change password", []*widget.FormItem{
		widget.NewFormItem("New password", password), widget.NewFormItem("Repeat password", confirm),
	}, "The affected player signs in again with the new password.", func() (operatorwire.Request, error) {
		if password.Text == "" || password.Text != confirm.Text {
			return operatorwire.Request{}, errors.New("enter matching passwords")
		}
		return operatorwire.Request{Action: "password", AccountID: a.ID, Password: password.Text}, nil
	}, func() { p.status.SetText("Password updated.") })
}

func (p *accountPanel) editAccount(a operatorwire.Account) {
	gm := widget.NewCheck("Game master", nil)
	gm.SetChecked(a.GM)
	points := widget.NewEntry()
	points.SetText(strconv.FormatUint(uint64(a.Points), 10))
	p.form("Role and points for "+a.Name, "Save", []*widget.FormItem{
		widget.NewFormItem("Role", gm), widget.NewFormItem("Item Mall points", points),
	}, "Points update the Item Mall balance live. A GM role change signs the affected player out.", func() (operatorwire.Request, error) {
		value, err := strconv.ParseUint(strings.TrimSpace(points.Text), 10, 31)
		if err != nil {
			return operatorwire.Request{}, errors.New("enter a whole point balance between 0 and 2,147,483,647")
		}
		return operatorwire.Request{Action: "account", AccountID: a.ID, GM: gm.Checked, Points: uint32(value), ExpectedAccount: &a}, nil
	}, func() { p.show(a.ID) })
}

func (p *accountPanel) editStats(a operatorwire.Account, ch operatorwire.Character) {
	req := operatorwire.Request{Action: "stats", AccountID: a.ID, CharacterID: ch.ID}
	p.request(req, func(res operatorwire.Response, err error) {
		if err != nil {
			return
		}
		old := *res.Stats
		values := []uint64{uint64(old.Strength), uint64(old.Dexterity), uint64(old.Rec), uint64(old.Intelligence), uint64(old.Wisdom), uint64(old.Luck),
			uint64(old.StatPoints), uint64(old.SkillPoints), uint64(old.Gold), uint64(old.EXP), uint64(old.Kills)}
		names := []string{"Strength", "Dexterity", "Rec", "Intelligence", "Wisdom", "Luck", "Unspent stat points", "Unspent skill points", "Gold", "Total EXP", "PvP rank kills"}
		limits := []uint64{32000, 32000, 32000, 32000, 32000, 32000, 32767, 32767, 4000000000, uint64(max(res.MaxEXP, old.EXP)), pvprank.CounterCap}
		items := make([]*widget.FormItem, len(names))
		entries := make([]*widget.Entry, len(names))
		for i, name := range names {
			entries[i] = widget.NewEntry()
			entries[i].SetText(strconv.FormatUint(values[i], 10))
			items[i] = widget.NewFormItem(name, entries[i])
		}
		status := "Offline"
		if res.Online {
			status = "Online"
		}
		help := fmt.Sprintf("%s. Total EXP at this level ranges from %d to %d. Reopen to refresh live values.", status, res.MinEXP, res.MaxEXP)
		p.form("Character stats for "+ch.Name, "Save stats", items, help, func() (operatorwire.Request, error) {
			for i, entry := range entries {
				value, err := strconv.ParseUint(strings.TrimSpace(entry.Text), 10, 32)
				if err != nil || value > limits[i] {
					return operatorwire.Request{}, fmt.Errorf("%s must be a whole number between 0 and %d", names[i], limits[i])
				}
				values[i] = value
			}
			next := &operatorwire.Stats{Strength: uint16(values[0]), Dexterity: uint16(values[1]), Rec: uint16(values[2]), Intelligence: uint16(values[3]),
				Wisdom: uint16(values[4]), Luck: uint16(values[5]), StatPoints: uint16(values[6]), SkillPoints: uint16(values[7]), Gold: uint32(values[8]), EXP: uint32(values[9]), Kills: uint32(values[10])}
			return operatorwire.Request{Action: "save-stats", AccountID: a.ID, CharacterID: ch.ID, ExpectedStats: &old, Stats: next}, nil
		}, func() { p.show(a.ID) })
	})
}

func (p *accountPanel) deleteCharacter(a operatorwire.Account, ch operatorwire.Character) {
	name := widget.NewEntry()
	p.form("Delete "+ch.Name, "Delete character", []*widget.FormItem{widget.NewFormItem("Type "+ch.Name, name)},
		"The character must be offline. GX allows restoration for two hours if the slot is free. Deletion removes guild membership.", func() (operatorwire.Request, error) {
			if name.Text != ch.Name {
				return operatorwire.Request{}, errors.New("type the exact character name")
			}
			return operatorwire.Request{Action: "delete", AccountID: a.ID, CharacterID: ch.ID, ConfirmName: name.Text}, nil
		}, func() { p.show(a.ID) })
}
