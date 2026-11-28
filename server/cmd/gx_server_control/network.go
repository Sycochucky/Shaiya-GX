package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/control"
)

func endpointWithPort(endpoint string, port int) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(endpoint))
	if err != nil || net.ParseIP(host).To4() == nil {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func (d *dashboard) networkView() fyne.CanvasObject {
	values, overridden, err := d.manager.NetworkSettings()
	if err != nil {
		return widget.NewLabel("Could not load network settings: " + err.Error())
	}
	login := widget.NewEntry()
	login.SetText(values.LoginListen)
	world := widget.NewEntry()
	world.SetText(values.WorldListen)
	publicIP := widget.NewEntry()
	publicIP.SetText(values.PublicIP)
	fields := []struct {
		entry *widget.Entry
		key   string
	}{
		{login, config.LoginListenKey},
		{world, config.WorldListenKey},
		{publicIP, config.PublicIPKey},
	}
	for _, field := range fields {
		if overridden[field.key] {
			field.entry.Disable()
		}
	}

	message := widget.NewLabel("Changes are saved only while Login and World are stopped.")
	message.Wrapping = fyne.TextWrapWord
	listenHost := func(entry *widget.Entry, key, host, defaultPort string) {
		if overridden[key] {
			return
		}
		_, port, _ := net.SplitHostPort(strings.TrimSpace(entry.Text))
		if port == "" {
			port = defaultPort
		}
		entry.SetText(net.JoinHostPort(host, port))
	}
	local := widget.NewButton("Local only", func() {
		listenHost(login, config.LoginListenKey, "127.0.0.1", "30800")
		listenHost(world, config.WorldListenKey, "127.0.0.1", "30810")
		if !overridden[config.PublicIPKey] {
			publicIP.SetText("127.0.0.1")
		}
		message.SetText("Local-only preset selected.")
	})
	public := widget.NewButton("Open for remote players", func() {
		listenHost(login, config.LoginListenKey, "0.0.0.0", "30800")
		listenHost(world, config.WorldListenKey, "0.0.0.0", "30810")
		message.SetText("Remote preset selected. Enter the IPv4 address players can reach.")
	})
	smartPorts := widget.NewButton("Find free ports", func() {
		loginPort, worldPort, err := control.SuggestPorts()
		if err != nil {
			message.SetText(err.Error())
			return
		}
		if !overridden[config.LoginListenKey] {
			login.SetText(endpointWithPort(login.Text, loginPort))
		}
		if !overridden[config.WorldListenKey] {
			world.SetText(endpointWithPort(world.Text, worldPort))
		}
		message.SetText(fmt.Sprintf("Selected available TCP ports: Login %d, World %d.", loginPort, worldPort))
	})

	addresses := control.LocalIPv4s()
	detected := widget.NewSelect(addresses, nil)
	if len(addresses) > 0 {
		detected.SetSelected(addresses[0])
		for _, address := range addresses {
			if address != "127.0.0.1" {
				detected.SetSelected(address)
				break
			}
		}
	}
	useDetected := widget.NewButton("Use selected IP", func() {
		if detected.Selected != "" && !overridden[config.PublicIPKey] {
			publicIP.SetText(detected.Selected)
		}
	})
	if overridden[config.PublicIPKey] {
		useDetected.Disable()
	}

	form := widget.NewForm(
		&widget.FormItem{Text: "Login listen", Widget: login, HintText: "IPv4:port. 0.0.0.0 accepts remote connections."},
		&widget.FormItem{Text: "World listen", Widget: world, HintText: "IPv4:port. Use a different port from Login."},
		&widget.FormItem{Text: "Advertised IPv4", Widget: publicIP, HintText: "Address sent to clients for World. Must be reachable by them."},
	)
	warning := widget.NewLabel("For public access, forward both TCP ports, allow them through Windows Firewall and use upstream DDoS protection.")
	warning.Wrapping = fyne.TextWrapWord
	save := widget.NewButton("Save network settings", nil)
	save.Importance = widget.HighImportance
	editable := !overridden[config.LoginListenKey] || !overridden[config.WorldListenKey] || !overridden[config.PublicIPKey]
	setEnabled(save, editable)
	contents := container.NewVBox(
		container.NewHBox(local, public, smartPorts),
		form,
		container.NewBorder(nil, nil, widget.NewLabel("Detected local IPv4"), useDetected, detected),
		warning,
		container.NewBorder(nil, nil, nil, save, message),
	)
	save.OnTapped = func() {
		save.Disable()
		message.SetText("Saving network settings…")
		next := control.NetworkSettings{LoginListen: login.Text, WorldListen: world.Text, PublicIP: publicIP.Text}
		go func() {
			err := d.manager.SaveNetworkSettings(d.ctx, next)
			if d.ctx.Err() != nil {
				return
			}
			fyne.Do(func() {
				setEnabled(save, editable)
				if err != nil {
					message.SetText(err.Error())
					return
				}
				message.SetText("Network settings saved. Start the services when ready.")
				d.revision.Add(1)
			})
		}()
	}
	return contents
}

func (d *dashboard) setupView() fyne.CanvasObject {
	d.selectedGame = widget.NewLabel("")
	d.selectedGame.Truncation = fyne.TextTruncateEllipsis
	d.refreshSelectedGame()
	game := panel(container.NewVBox(strong("Game client"),
		container.NewBorder(nil, nil, nil, d.chooseGame, d.selectedGame)))
	server := panel(container.NewBorder(nil, nil, strong("Server installation"), d.servers, nil))
	network := panel(container.NewVBox(strong("Network"), d.networkView()))
	return container.NewVScroll(container.NewVBox(server, game, network))
}

func (d *dashboard) refreshSelectedGame() {
	if d.selectedGame == nil {
		return
	}
	path := gamePath(d.manager.Root())
	if path == "" {
		path = "No game client selected."
	}
	d.selectedGame.SetText(path)
}
