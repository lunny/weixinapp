package weixinapp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	accessURL = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
)

// APP defines a weixin app
type APP struct {
	appID       string
	appSecret   string
	accessToken string
	accessLock  sync.RWMutex
	expireAt    time.Time
}

// NewAPP creates a new weixin app
func NewAPP(appID, appSecret string) *APP {
	return &APP{
		appID:     appID,
		appSecret: appSecret,
	}
}

// AccessTokenResponse request access token response
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpireIn    int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// GetAccessToken get the access token
func (app *APP) GetAccessToken() (string, error) {
	app.accessLock.Lock()
	defer app.accessLock.Unlock()

	if len(app.accessToken) > 0 && time.Now().Before(app.expireAt.Add(-time.Minute*5)) {
		return app.accessToken, nil
	}

	return app.refreshAccessToken()
}

// RefreshAccessToken refresh accessToken
func (app *APP) RefreshAccessToken() (string, error) {
	app.accessLock.Lock()
	defer app.accessLock.Unlock()
	return app.refreshAccessToken()
}

func (app *APP) refreshAccessToken() (string, error) {
	resp, err := http.Get(fmt.Sprintf(accessURL, app.appID, app.appSecret))
	if err != nil {
		return "", errors.Wrap(err, "RequestAccessToken:Get")
	}
	defer resp.Body.Close()

	var result AccessTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", errors.Wrap(err, "RequestAccessToken:Decode")
	}

	if result.ErrCode > 0 {
		return "", fmt.Errorf("RequestAccessToken: %d %s", result.ErrCode, result.ErrMsg)
	}

	app.accessToken = result.AccessToken
	app.expireAt = time.Now().Add(time.Minute * time.Duration(result.ExpireIn))

	return app.accessToken, nil
}

const (
	createQRURL = "https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=%s"
)

func (app *APP) CreateQRCode(path string, width int, w io.Writer) error {
	accessToken, err := app.GetAccessToken()
	if err != nil {
		return err
	}

	strJSON := fmt.Sprintf(`{"path": "%s", "width": %d}`, path, width)

	resp, err := http.Post(fmt.Sprintf(createQRURL, accessToken),
		"application/x-www-form-urlencoded",
		strings.NewReader(strJSON))
	if err != nil {
		return errors.Wrap(err, "CreateQRCode POST")
	}

	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return errors.Wrap(err, "CreateQRCode Read Resp")
	}

	return nil
}
