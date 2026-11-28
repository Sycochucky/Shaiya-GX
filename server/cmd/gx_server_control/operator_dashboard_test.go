package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
	"github.com/Spelunkern/Shaiya-GX/server/internal/filelock"
	"github.com/Spelunkern/Shaiya-GX/server/internal/operatorwire"
)

func TestWindowsAccountsDashboard(t *testing.T) {
	root := os.Getenv("GX_NATIVE_OPERATOR_FIXTURE")
	if root == "" {
		t.Skip("requires completed disposable operator fixture")
	}
	app := test.NewApp()
	window := app.NewWindow("GX Accounts test")
	window.Resize(fyne.NewSize(1120, 760))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	m := control.New(root)
	if err := m.Apply(ctx, control.Start, control.Services(), nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Apply(context.Background(), control.ForceStop, control.Services(), nil); err != nil {
			t.Error(err)
		}
	})
	original := m.Snapshot(ctx)
	d := &dashboard{window: window, manager: m, ctx: ctx, rows: map[control.Service]serviceRow{}, status: map[control.Service]control.Status{}}
	d.build()
	d.checked = true
	d.status = m.Snapshot(ctx)
	d.refresh()
	wait := func(description string, find func() bool) {
		t.Helper()
		until := time.Now().Add(12 * time.Second)
		for time.Now().Before(until) {
			if find() {
				return
			}
			time.Sleep(70 * time.Millisecond)
		}
		t.Fatalf("missing %s", description)
	}
	top := func() fyne.CanvasObject { return window.Canvas().Overlays().Top() }
	var panel fyne.CanvasObject
	findButton := func(label string) *widget.Button {
		scope := top()
		if scope == nil {
			scope = panel
		}
		n := accountControl(scope, func(o fyne.CanvasObject) bool { b, ok := o.(*widget.Button); return ok && b.Text == label })
		if n == nil {
			return nil
		}
		return n.(*widget.Button)
	}
	button := func(label string) {
		t.Helper()
		var b *widget.Button
		wait(label, func() bool { b = findButton(label); return b != nil && !b.Disabled() })
		test.Tap(b)
	}
	fields := func() []*widget.FormItem {
		t.Helper()
		var n fyne.CanvasObject
		wait("operator form", func() bool {
			n = accountControl(top(), func(o fyne.CanvasObject) bool { _, ok := o.(*widget.Form); return ok })
			return n != nil
		})
		return n.(*widget.Form).Items
	}
	set := func(items []*widget.FormItem, i int, s string) { items[i].Widget.(*widget.Entry).SetText(s) }
	readAccount := func(name string) (operatorwire.Account, bool) {
		r, err := m.Operator(ctx, operatorwire.Request{Action: "accounts", Query: name})
		if err != nil || len(r.Accounts) != 1 {
			return operatorwire.Account{}, false
		}
		return r.Accounts[0], true
	}
	search := func(name string) {
		t.Helper()
		n := accountControl(panel, func(o fyne.CanvasObject) bool {
			e, ok := o.(*widget.Entry)
			return ok && e.PlaceHolder == "Search accounts by username"
		})
		l := accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok }).(*widget.List)
		e := n.(*widget.Entry)
		e.SetText(name)
		e.OnSubmitted(name)
		wait("one matching account", func() bool {
			if l.Length() != 1 {
				return false
			}
			label := widget.NewLabel("")
			l.UpdateItem(0, label)
			return label.Text == name
		})
		l.Select(0)
	}

	d.tabs.Select(d.accountsItem)
	accountTabs := d.accountsItem.Content.(*container.AppTabs)
	panel = accountTabs.Items[0].Content
	wait("initial account list", func() bool {
		l := accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok })
		return l != nil && l.(*widget.List).Length() > 0
	})
	name := fmt.Sprintf("GXUI%08d", time.Now().UnixNano()%100000000)
	button("New account")
	form := fields()
	set(form, 0, name)
	set(form, 1, "test GUI only password")
	set(form, 2, "test GUI only password")
	button("Create")
	wait("saved account", func() bool { _, ok := readAccount(name); return ok && top() == nil })
	wait("creation search resolved", func() bool {
		l := accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok }).(*widget.List)
		return l.Length() == 1
	})
	accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok }).(*widget.List).Select(0)
	button("Role and points")
	form = fields()
	form[0].Widget.(*widget.Check).SetChecked(true)
	set(form, 1, "420")
	button("Save")
	wait("GM and point balance from GUI", func() bool { a, ok := readAccount(name); return ok && a.GM && a.Points == 420 && top() == nil })
	button("Change password")
	form = fields()
	set(form, 0, "reset GUI only password")
	set(form, 1, "reset GUI only password")
	button("Change password")
	wait("password dialog closed", func() bool { return top() == nil })
	button("Ban account")
	form = fields()
	set(form, 0, "disposable QA")
	button("Ban account")
	wait("account ban", func() bool {
		a, ok := readAccount(name)
		return ok && a.Banned && a.BanReason == "disposable QA" && findButton("Unban account") != nil
	})
	button("Unban account")
	button("Unban")
	wait("account unban", func() bool {
		a, ok := readAccount(name)
		return ok && !a.Banned && findButton("Ban account") != nil
	})

	accountTabs.SelectIndex(1)
	panel = accountTabs.Items[1].Content
	button("Ban IP…")
	form = fields()
	set(form, 0, "192.0.2.123")
	set(form, 1, "disposable IP QA")
	button("Ban IP")
	ipList := accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok }).(*widget.List)
	wait("IP listed", func() bool { return ipList.Length() == 1 })
	ipList.Select(0)
	button("Unban IP")
	button("Unban")
	wait("removed IP", func() bool { return ipList.Length() == 0 })
	accountTabs.SelectIndex(0)
	panel = accountTabs.Items[0].Content

	search("GXProbeOperator")
	button("Edit stats")
	form = fields()
	set(form, 0, "444")
	set(form, 10, "67")
	button("Save stats")
	wait("saved GUI stats", func() bool {
		a, ok := readAccount("GXProbeOperator")
		if !ok {
			return false
		}
		detail, err := m.Operator(ctx, operatorwire.Request{Action: "details", AccountID: a.ID})
		if err != nil || len(detail.Characters) != 1 {
			return false
		}
		r, err := m.Operator(ctx, operatorwire.Request{Action: "stats", AccountID: a.ID, CharacterID: detail.Characters[0].ID})
		return err == nil && r.Stats.Strength == 444 && r.Stats.Kills == 67 && top() == nil
	})
	for service, current := range m.Snapshot(ctx) {
		if current.State != control.Ready || current.PID != original[service].PID {
			t.Fatalf("%s was interrupted during GUI edits: %+v", service, current)
		}
	}
	button("Delete")
	form = fields()
	set(form, 0, "GXProbeChar")
	button("Delete character")
	wait("restore action for deleted character", func() bool { return findButton("Restore") != nil })
	button("Restore")
	wait("restored character actions", func() bool { return findButton("Edit stats") != nil })
	for service, current := range m.Snapshot(ctx) {
		if current.State != control.Ready || current.PID != original[service].PID {
			t.Fatalf("%s was interrupted by GUI delete/restore: %+v", service, current)
		}
	}
	if err := m.Apply(ctx, control.Stop, control.Services(), nil); err != nil {
		t.Fatal(err)
	}
	t.Log("GUI forms saved account and stat changes, then deleted and restored an offline character with both services running.")
}

func TestAccountSearchKeepsLatestSubmission(t *testing.T) {
	root := os.Getenv("GX_NATIVE_OPERATOR_FIXTURE")
	if root == "" {
		t.Skip("requires a disposable fixture with GXProbeOperator")
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	app := test.NewApp()
	window := app.NewWindow("GX search test")
	defer window.Close()
	manager := control.New(root)
	d := &dashboard{window: window, manager: manager, ctx: ctx, rows: map[control.Service]serviceRow{}, status: map[control.Service]control.Status{}}
	d.build()

	// Delay the automatic query so both submissions happen while it is pending.
	release, err := filelock.Acquire(ctx, filepath.Join(root, "runtime", "control", "operations.lock"))
	if err != nil {
		t.Fatal(err)
	}
	d.tabs.Select(d.accountsItem)
	panel := d.accountsItem.Content.(*container.AppTabs).Items[0].Content
	input := accountControl(panel, func(o fyne.CanvasObject) bool {
		e, ok := o.(*widget.Entry)
		return ok && e.PlaceHolder == "Search accounts by username"
	})
	if input == nil {
		release()
		t.Fatal("account search field unavailable")
	}
	entry := input.(*widget.Entry)
	for _, query := range []string{"missing-first-query", "GXProbeOperator"} {
		entry.SetText(query)
		entry.OnSubmitted(query)
	}
	release()

	l := accountControl(panel, func(o fyne.CanvasObject) bool { _, ok := o.(*widget.List); return ok }).(*widget.List)
	for deadline := time.Now().Add(25 * time.Second); time.Now().Before(deadline); {
		if l.Length() == 1 {
			label := widget.NewLabel("")
			l.UpdateItem(0, label)
			if label.Text == "GXProbeOperator" {
				return
			}
		}
		time.Sleep(70 * time.Millisecond)
	}
	t.Fatal("latest submitted account search was not displayed")
}

func accountControl(root fyne.CanvasObject, matches func(fyne.CanvasObject) bool) fyne.CanvasObject {
	visited := map[fyne.CanvasObject]bool{}
	var scan func(fyne.CanvasObject) fyne.CanvasObject
	scan = func(object fyne.CanvasObject) fyne.CanvasObject {
		if object == nil || visited[object] {
			return nil
		}
		visited[object] = true
		if matches(object) {
			return object
		}
		var children []fyne.CanvasObject
		switch x := object.(type) {
		case *fyne.Container:
			children = x.Objects
		case fyne.Widget:
			children = test.WidgetRenderer(x).Objects()
		}
		for _, child := range children {
			if m := scan(child); m != nil {
				return m
			}
		}
		return nil
	}
	return scan(root)
}
