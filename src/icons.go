package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/Vyrnexis/welcome-app/assets"
)

// embeddedResource loads an embedded asset as a Fyne resource, or nil when it does not exist.
func embeddedResource(name string) fyne.Resource {
	data, err := assets.FS.ReadFile(name)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

// desktopIcon shows the embedded logo for the detected desktop, or a generic icon when there is none.
func desktopIcon(desktop DesktopInfo) fyne.CanvasObject {
	if resource := embeddedResource(desktop.Key + ".svg"); resource != nil {
		icon := canvas.NewImageFromResource(resource)
		icon.SetMinSize(fyne.NewSize(54, 54))
		icon.FillMode = canvas.ImageFillContain
		return icon
	}

	icon := widget.NewIcon(theme.ComputerIcon())
	return minSize(container.NewCenter(icon), fyne.NewSize(54, 54))
}
