package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"

	commands "github.com/cardinalby/vlc-sync-play/build/internal"
	"github.com/cardinalby/vlc-sync-play/build/internal/consts"
	"github.com/cardinalby/vlc-sync-play/build/internal/macos"
)

var rootPath = getRootPath()
var distDir = path.Join(rootPath, "dist")
var tmpDir = path.Join(distDir, "tmp")

var srcPsdIconPath = path.Join(rootPath, "assets/icons/icon.psd")
var winManifestPath = path.Join(rootPath, "assets/manifest/exe.manifest")
var plistManifestPath = path.Join(rootPath, "assets/manifest/Info.plist")

const macosBundleFileName = "VLC Sync Play.app"

var macosBundles = map[consts.Arch]macos.BundlePath{
	consts.ArchAmd64: macos.BundlePath(path.Join(distDir, "macos_amd64", macosBundleFileName)),
	consts.ArchArm64: macos.BundlePath(path.Join(distDir, "macos_arm64", macosBundleFileName)),
}

var dstWinSysoPath = path.Join(rootPath, "cmd/internal/rsrc_windows_amd64.syso")
var iconSetDstFileName = "icon.icns"

// var trayAgentMainPkgRelPath = path.Join("cmd/trayagent")
var trayAgentMainPkgRelPath = path.Join("cmd/cliagent")

const binFileName = "vlc-sync-play"

var windowsAmd64BinPath = path.Join(distDir, "windows_amd64", binFileName+".exe")
var macosAmd64BinPath = path.Join(macosBundles[consts.ArchAmd64].GetBinDir(), binFileName)
var macosArm64BinPath = path.Join(macosBundles[consts.ArchArm64].GetBinDir(), binFileName)
var linuxAmd64BinPath = path.Join(distDir, "linux_amd64", binFileName)
var linuxArm64BinPath = path.Join(distDir, "linux_arm64", binFileName)

func getCmd(logger *log.Logger) cmd {
	return sequential{
		"win_syso": cmdFunc(func() error {
			return commands.GenerateWindowsSysoFile(
				srcPsdIconPath, winManifestPath, tmpDir, dstWinSysoPath, logger,
			)
		}),
		"go_build": cmdFunc(func() error {
			return commands.GoBuild(
				rootPath,
				trayAgentMainPkgRelPath,
				map[commands.Target]string{
					//{Os: consts.OsWindows, Arch: consts.ArchAmd64}: windowsAmd64BinPath,
					//{Os: consts.OsDarwin, Arch: consts.ArchAmd64}:  macosAmd64BinPath,
					//{Os: consts.OsDarwin, Arch: consts.ArchArm64}:  macosArm64BinPath,
					//{Os: consts.OsLinux, Arch: consts.ArchAmd64}:   linuxAmd64BinPath,
					{Os: consts.OsLinux, Arch: consts.ArchArm64}: linuxArm64BinPath,
				},
				tmpDir,
				logger,
			)
		}),
		"mac_bundles": cmdFunc(func() error {
			return commands.PackMacosBundles(
				macosBundles,
				srcPsdIconPath,
				plistManifestPath,
				iconSetDstFileName,
				tmpDir,
			)
		}),
	}
}

func getLogger() *log.Logger {
	isDebug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()
	if *isDebug {
		return log.New(os.Stdout, "", 1)
	} else {
		return log.New(io.Discard, "", 0)
	}
}

func main() {
	logger := getLogger()

	err := getCmd(logger).run()
	if err := os.RemoveAll(tmpDir); err != nil {
		logger.Printf("error removing tmp '%s': %v", tmpDir, err)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
