package app

import (
	"context"

	"github.com/cardinalby/vlc-sync-play/internal/cliagent/args"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/protocols"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/instance"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/syncer"
)

type CliAgentApp struct {
	args        args.CmdLineArgs
	apiProtocol protocols.ApiProtocol
}

func NewCliAgentApp(args args.CmdLineArgs) *CliAgentApp {
	return &CliAgentApp{
		args:        args,
		apiProtocol: protocols.ApiProtocolHttpJson,
	}
}

func (a *CliAgentApp) Start(ctx context.Context) error {
	instanceLauncher := instance.GetLauncher(a.args.VlcPath, a.apiProtocol)
	masterInstance, err := instanceLauncher(a.args.FilePaths, false)
	if err != nil {
		return err
	}
	playersSyncer := syncer.NewSyncer(
		masterInstance,
		a.args.GetPollingInterval(),
		a.args.InstancesNumber,
		instanceLauncher,
	)
	return playersSyncer.Start(ctx)
}
