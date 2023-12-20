package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/cardinalby/vlc-sync-play/internal/cliagent/app"
	"github.com/cardinalby/vlc-sync-play/internal/cliagent/args"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cmdLineArgs, err := args.ParseCmdLineArgs()
	if err != nil {
		panic(err)
	}
	application := app.NewCliAgentApp(cmdLineArgs)
	ctx := getAppContext()
	err = application.Start(ctx)
	if err != nil && !errors.Is(err, ctx.Err()) {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}

func getAppContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	call := make(chan os.Signal, 1)
	signal.Notify(call, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-call
		cancel()
	}()
	return ctx
}
