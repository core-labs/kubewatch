package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitnami-labs/kubewatch/config"
	"net/http"
	"os"
	uri "net/url"
	"strconv"
	"time"
)

var dingtalkErrMsg = `
%s

You need to set dingtalk url, secret for Dingtalk notify,
using "--url/-u" and "--secret/-s", or using environment variables:

export KW_DINGTALK_URL=dingtalk_url
export KW_DINGTALK_SECRET=dingtalk_secret

Command line flags will override environment variables

`

type DingTalk struct {
	Url    string
	Secret string
}

func (d *DingTalk) Init(c *config.Config) error {
	url := c.Handler.DingTalk.Url
	secret := c.Handler.DingTalk.Secret
	if url == "" {
		url = os.Getenv("KW_DINGTALK_URL")
	}

	if secret == "" {
		secret = os.Getenv("KW_DINGTALK_SECRET")
	}

	d.Secret = secret
	d.Url = url

	return checkMissingDingTalkVars(d)

}

func (d *DingTalk) ObjectCreated(obj interface{}) {
	panic("implement me")
}

func (d *DingTalk) ObjectDeleted(obj interface{}) {
	panic("implement me")
}

func (d *DingTalk) ObjectUpdated(oldObj, newObj interface{}) {
	panic("implement me")
}

func (d *DingTalk) TestHandler() {
	panic("implement me")
}

type DingtalkMessage struct {
}

func checkMissingDingTalkVars(talk *DingTalk) error {

	if talk.Url == "" || talk.Secret == "" {
		return fmt.Errorf(dingtalkErrMsg, "Missing DingTalk url or secret")
	}

	return nil

}
func postMessage(url string, secret string, msg *DingtalkMessage) error {

	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	u, err := uri.Parse(url)
	if err != nil {
		return err
	}

	t := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)

	qs := u.Query()
	qs.Set("timestamp", t)
	qs.Set("sign", SignRequestData(t, secret))

	req, err := http.NewRequest("POST", qs.Encode(), bytes.NewBuffer(message))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil

}

func hmacSha256(data string, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return h.Sum(nil)
}

func SignRequestData(t string, secret string) string {
	string2sign := fmt.Sprintf("%s\n%s", t, secret)
	return base64.StdEncoding.EncodeToString(hmacSha256(string2sign, secret))
}
