//go:build darwin

package assets

import "embed"

//go:embed icons/template_tray_icon.png
//go:embed generated/tray_icon.png
var IconFS embed.FS

func GetTrayIcon() []byte {
	data, err := IconFS.ReadFile("generated/tray_icon.png")
	if err != nil {
		panic(err)
	}
	return data
}

func GetTemplateTrayIcon() []byte {
	data, err := IconFS.ReadFile("icons/template_tray_icon.png")
	if err != nil {
		return nil
	}
	return data
}
