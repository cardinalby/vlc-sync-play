package main

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
)

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func getRootPath() (res string) {
	defer func() {
		doModFilePath := filepath.Join(res, "go.mod")
		f, err := os.Open(doModFilePath)
		if err != nil {
			panic(doModFilePath + " not found")
		}
		defer func() {
			_ = f.Close()
		}()
		scanner := bufio.NewScanner(f)
		if !scanner.Scan() {
			panic(doModFilePath + " is empty")
		}
		if scanner.Text() != "module github.com/cardinalby/vlc-sync-play" {
			panic(doModFilePath + " has wrong module name: " + scanner.Text())
		}
	}()
	if vlcspRoot, ok := os.LookupEnv("VLCSP_ROOT"); ok {
		return path.Join(getWd(), vlcspRoot)
	}
	return path.Join(getWd(), "..")
}
