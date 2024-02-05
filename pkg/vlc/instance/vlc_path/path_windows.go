package vlc_path

import "golang.org/x/sys/windows/registry"

func getDefaultVlcBinPath() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\VideoLAN\VLC`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	strVal, _, err := key.GetStringValue("")
	if err != nil {
		return "", err
	}
	return strVal, nil
}
