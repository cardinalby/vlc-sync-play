package protocols

import (
	"errors"

	"github.com/cardinalby/vlc-sync-play/pkg/util/logging"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic/httpjson"
)

const ApiProtocolHttpJson ApiProtocol = "http-json"

type ApiProtocol string

var ErrUnsupportedApiProtocol = errors.New("unsupported api protocol")

func (p ApiProtocol) Validate() error {
	if p != ApiProtocolHttpJson {
		return ErrUnsupportedApiProtocol
	}
	return nil
}

func NewLocalBasicApiClient(protocol ApiProtocol, logger logging.Logger) (basic.ApiClient, error) {
	switch protocol {
	case ApiProtocolHttpJson:
		return httpjson.NewLocalBasicApiClient(logger)
	default:
		return nil, ErrUnsupportedApiProtocol
	}
}
