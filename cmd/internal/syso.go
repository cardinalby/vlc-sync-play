package internal

func init() {
	// this file is required to make go build link generated rsrc_windows_amd64.syso file
	// Build can be performed without this file, but it contains an icon and a manifest
}
