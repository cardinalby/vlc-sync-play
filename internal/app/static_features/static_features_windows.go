//go:build windows

package static_features

// Can't intercept stderr
const ClickPause = false

// Launching with file requires manual setup of VLC ("Allow only one instance when started from file" option)
const LaunchWithFile = false
