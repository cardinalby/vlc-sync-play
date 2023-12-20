package httpjson

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	osutil "github.com/cardinalby/vlc-sync-play/pkg/util/os"
	rndutil "github.com/cardinalby/vlc-sync-play/pkg/util/rnd"
	timeutil "github.com/cardinalby/vlc-sync-play/pkg/util/time"
	"github.com/cardinalby/vlc-sync-play/pkg/vlc/client/basic"
)

var ErrFormingRequest = errors.New("failed to form a request")
var ErrAuthFailed = errors.New("authentication failed")

const passwordLength = 6
const apiEndpointStatus = "status.json"
const apiEndpointPlaylist = "playlist.json"

// BasicApiClient is a client for the VLC HTTP JSON API.
// See: https://github.com/videolan/vlc/tree/master/share/lua/http/requests
type BasicApiClient struct {
	connectionInfo ConnectionInfo
	baseUrl        string
	authHeader     string
}

func NewLocalBasicApiClient() (*BasicApiClient, error) {
	password := rndutil.GeneratePassword(passwordLength)
	host, port, err := osutil.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %w", err)
	}
	return NewBasicApiClient(ConnectionInfo{
		Host:     host,
		Port:     port,
		Password: password,
	}), nil
}

func NewBasicApiClient(connectionInfo ConnectionInfo) *BasicApiClient {
	//goland:noinspection HttpUrlsUsage
	return &BasicApiClient{
		connectionInfo: connectionInfo,
		baseUrl:        fmt.Sprintf("http://%s:%d/requests/", connectionInfo.Host, connectionInfo.Port),
		authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte(":"+connectionInfo.Password)),
	}
}

func (apiClient *BasicApiClient) GetStatus(ctx context.Context) (basic.Status, error) {
	statusDto, moment, err := sendApiRequest[statusDto](
		ctx,
		apiClient,
		apiEndpointStatus,
		nil,
	)
	if err != nil {
		return basic.Status{}, err
	}
	return toStatus(statusDto, moment), nil
}

func (apiClient *BasicApiClient) SendStatusCmd(ctx context.Context, cmd basic.Command) (basic.Status, error) {
	fmt.Printf("%s SendStatusCmd: %v\n", apiClient.baseUrl, cmd)
	statusDto, moment, err := sendApiRequest[statusDto](
		ctx,
		apiClient,
		apiEndpointStatus,
		cmd,
	)
	if err != nil {
		return basic.Status{}, err
	}
	return toStatus(statusDto, moment), nil
}

func (apiClient *BasicApiClient) GetCurrentFileUri(ctx context.Context) (string, error) {
	playlistItemDto, _, err := sendApiRequest[playlistItemDto](
		ctx,
		apiClient,
		apiEndpointPlaylist,
		nil,
	)
	if err != nil {
		return "", err
	}
	currentItem, hasCurrent := playlistItemDto.getCurrent()
	if !hasCurrent {
		return "", nil
	}
	return currentItem.Uri, nil
}

func (apiClient *BasicApiClient) IsRecoverableErr(err error) bool {
	return !(errors.Is(err, ErrFormingRequest) && !errors.Is(err, ErrAuthFailed))
}

func (apiClient *BasicApiClient) GetLaunchArgs() []string {
	return []string{
		"--extraintf=http",
		"--http-host",
		apiClient.connectionInfo.Host,
		"--http-port",
		strconv.Itoa(apiClient.connectionInfo.Port),
		"--http-password",
		apiClient.connectionInfo.Password,
	}
}

func sendApiRequest[T any](
	ctx context.Context,
	apiClient *BasicApiClient,
	endpoint string,
	cmd map[basic.Key]string,
) (response T, moment timeutil.Range, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiClient.baseUrl+endpoint, nil)
	if err != nil {
		return response, moment, fmt.Errorf("%w: %s", ErrFormingRequest, err.Error())
	}
	req.Header.Add("Authorization", apiClient.authHeader)

	query := req.URL.Query()
	for name, val := range cmd {
		query.Add(string(name), val)
	}
	req.URL.RawQuery = query.Encode()

	moment.Min = time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response, moment, err
	}
	moment.Max = time.Now()

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 401 {
			return response, moment, ErrAuthFailed
		}
		return response, moment, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, moment, fmt.Errorf("error reading response body: %w", err)
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return response, moment, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return response, moment, nil
}

func toStatus(dto statusDto, moment timeutil.Range) basic.Status {
	return basic.Status{
		Moment:    moment,
		LengthSec: dto.LengthSec,
		Rate:      dto.Rate,
		State:     dto.State,
		Position:  dto.Position,
		FileName:  dto.getFileName(),
	}
}
