package protocols

import (
	"errors"

	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/httpjson"
)

const ApiProtocolHttpJson ApiProtocol = "http-json"

type ApiProtocol string

func NewLocalBasicApiClient(protocol ApiProtocol) (basic.ApiClient, error) {
	switch protocol {
	case ApiProtocolHttpJson:
		return httpjson.NewLocalBasicApiClient()
	default:
		return nil, errors.New("unsupported api protocol")
	}
}
