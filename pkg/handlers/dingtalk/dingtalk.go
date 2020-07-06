package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bitnami-labs/kubewatch/config"
	kbEvent "github.com/bitnami-labs/kubewatch/pkg/event"
	"log"
	"net/http"
	uri "net/url"
	"os"
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

type DingTalkMessage struct {
	Type     string                  `json:"msgtype"`
	Markdown DingTalkMarkdownMessage `json:"markdown"`
}

type DingTalkMarkdownMessage struct {
	Title string `json:"title"`
	Text  string `json:"text"`
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
	notifyDingTalk(d, obj, "created")
}

func (d *DingTalk) ObjectDeleted(obj interface{}) {
	notifyDingTalk(d, obj, "deleted")
}

func (d *DingTalk) ObjectUpdated(oldObj, newObj interface{}) {
	notifyDingTalk(d, newObj, "updated")
}

func (d *DingTalk) TestHandler() {
	message := &DingTalkMessage{
		Type: "markdown",
		Markdown: DingTalkMarkdownMessage{
			Title: "测试资源更新",
			Text:  fmt.Sprintf("#### Kubewatch 测试 \n> 命名空间 **`%s`** 中的 **`%s`** 已经被 **`%s`**:\n**`%s`**", "default", "Deployment", "Created", "test"),
		},
	}

	err := postMessage(d.Url, d.Secret, message)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	log.Printf("Message successfully sent to url %s at %s", d.Url, time.Now())
}

func checkMissingDingTalkVars(talk *DingTalk) error {

	if talk.Url == "" || talk.Secret == "" {
		return fmt.Errorf(dingtalkErrMsg, "Missing DingTalk url or secret")
	}
	return nil
}

func notifyDingTalk(d *DingTalk, obj interface{}, action string) {
	e := kbEvent.New(obj, action)

	dingtalkMessage := prepareDingtalkMessage(e, d)

	err := postMessage(d.Url, d.Secret, dingtalkMessage)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	log.Printf("Message successfully sent to channel %s at %s", d.Url, time.Now())
}

func prepareDingtalkMessage(e kbEvent.Event, m *DingTalk) *DingTalkMessage {

	return &DingTalkMessage{
		Type: "markdown",
		Markdown: DingTalkMarkdownMessage{
			Title: fmt.Sprintf("%s 被 %s", e.Kind, e.Status),
			Text:  eventI18n(e),
		},
	}
}

func eventI18n(e kbEvent.Event) string {
	var msg string
	switch e.Kind {
	case "namespace":
		msg = fmt.Sprintf(
			"> 命名空间 `%s` 已经被 `%s`",
			e.Name,
			e.Reason,
		)
	default:
		msg = fmt.Sprintf(
			"> 命名空间 `%s` 中的 `%s` 已经被 `%s`:\n`%s`",
			e.Namespace,
			e.Kind,
			e.Reason,
			e.Name,
		)
	}
	return msg
}

func postMessage(url string, secret string, msg *DingTalkMessage) error {

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

	u.RawQuery = qs.Encode()

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(message))
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
