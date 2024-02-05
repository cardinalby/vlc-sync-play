//go:build windows

package assets

import "embed"

//go:embed generated/tray_icon.ico
var IconFS embed.FS

func GetTrayIcon() []byte {
	data, err := IconFS.ReadFile("generated/tray_icon.ico")
	if err != nil {
		panic(err)
	}
	return data
}

func GetTemplateTrayIcon() []byte {
	return nil
}
