package main

import (
	"context"
	"fmt"

	_ "github.com/cardinalby/vlc-sync-play/cmd/internal"
	"github.com/cardinalby/vlc-sync-play/internal/cli"
)

func main() {
	if err := cli.RunCliApp(context.Background()); err != nil {
		fmt.Printf(err.Error())
	}
}
