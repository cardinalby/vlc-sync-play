package basic

import (
	"context"
)

type ApiClient interface {
	GetStatus(ctx context.Context) (Status, error)
	SendStatusCmd(ctx context.Context, cmd Command) (Status, error)
	GetCurrentFileUri(ctx context.Context) (string, error)
	IsRecoverableErr(err error) bool
	GetLaunchArgs() []string
}
