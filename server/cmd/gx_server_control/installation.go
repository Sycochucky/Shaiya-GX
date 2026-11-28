package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

func installationPathFile(root string) string {
	return filepath.Join(root, ".local", "control-installation")
}

func installationListFile(root string) string {
	return filepath.Join(root, ".local", "control-installations.json")
}

func rememberInstallation(root, chosen string) error {
	directory := filepath.Join(root, ".local")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	chosen = filepath.Clean(chosen)
	if err := os.WriteFile(installationPathFile(root), []byte(chosen), 0600); err != nil {
		return err
	}
	installations := rememberedInstallations(root)
	found := false
	for _, existing := range installations {
		if control.SameInstallation(existing, chosen) {
			found = true
			break
		}
	}
	if !found {
		installations = append(installations, chosen)
	}
	sort.Strings(installations)
	data, err := json.MarshalIndent(installations, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(installationListFile(root), append(data, '\n'), 0600)
}

func rememberedInstallations(root string) []string {
	data, err := os.ReadFile(installationListFile(root))
	if err != nil {
		return nil
	}
	var stored []string
	if json.Unmarshal(data, &stored) != nil {
		return nil
	}
	var result []string
	for _, path := range stored {
		path = strings.Trim(strings.TrimSpace(path), "\"")
		if !filepath.IsAbs(path) {
			continue
		}
		duplicate := false
		for _, existing := range result {
			if control.SameInstallation(existing, path) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, filepath.Clean(path))
		}
	}
	return result
}

func openInstallation(path string) (*control.Manager, error) {
	path = strings.Trim(strings.TrimSpace(path), "\"")
	if !filepath.IsAbs(path) {
		return nil, errors.New("enter the full path to a GX Server folder")
	}
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	owner := control.New(path)
	if err := owner.CheckInstallation(); err != nil {
		return nil, err
	}
	return owner, nil
}

func rememberedInstallation(original *control.Manager) *control.Manager {
	if original.CheckInstallation() == nil {
		return original
	}
	stored, err := os.ReadFile(installationPathFile(original.Root()))
	if err == nil {
		if owner, err := openInstallation(string(stored)); err == nil {
			return owner
		}
	}
	return original
}

func (d *dashboard) chooseInstallation() {
	if d.useInstallation == nil {
		return
	}
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Full path to the GX Server folder")
	message := widget.NewLabel("Choose an isolated GX installation with Login, World, and GX Console. Switching dashboards does not stop another installation.")
	message.Wrapping = fyne.TextWrapWord
	open := widget.NewButton("Open Dashboard", nil)
	open.Importance = widget.HighImportance
	choices := append([]string{d.manager.Root()}, rememberedInstallations(d.controllerRoot)...)
	choices = append(choices, d.availableOwners...)
	var unique []string
	for _, candidate := range choices {
		duplicate := false
		for _, existing := range unique {
			if control.SameInstallation(existing, candidate) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, candidate)
		}
	}
	contents := container.NewVBox(entry, message, open)
	if len(unique) > 0 {
		list := widget.NewList(func() int { return len(unique) },
			func() fyne.CanvasObject {
				label := widget.NewLabel("")
				label.Truncation = fyne.TextTruncateEllipsis
				return label
			},
			func(i widget.ListItemID, object fyne.CanvasObject) {
				object.(*widget.Label).SetText(unique[i])
			})
		list.OnSelected = func(i widget.ListItemID) { entry.SetText(unique[i]) }
		contents = container.NewVBox(widget.NewLabel("Known and verified running installations"),
			container.NewGridWrap(fyne.NewSize(720, float32(min(len(unique), 4)*36)), list), contents)
		entry.SetText(d.manager.Root())
	}
	picker := dialog.NewCustom("GX Server Installations", "Cancel", contents, d.window)
	picker.Resize(fyne.NewSize(760, 300))
	open.OnTapped = func() {
		owner, err := openInstallation(entry.Text)
		if err != nil {
			message.SetText(err.Error())
			return
		}
		picker.Hide()
		d.useInstallation(owner)
	}
	entry.OnSubmitted = func(string) { open.OnTapped() }
	picker.Show()
}
