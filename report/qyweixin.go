package report

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type QYWeiXinReport struct {
	endpoint    string
	accessToken string
	client      http.Client
	limiter     *rate.Limiter
}

type QYWeiXinReportFunc func(report *QYWeiXinReport)

func WithReportLimiter(limiter *rate.Limiter) QYWeiXinReportFunc {
	return func(report *QYWeiXinReport) {
		report.limiter = limiter
	}
}

func NewQYWeiXinReport(token string, opts ...QYWeiXinReportFunc) *QYWeiXinReport {
	r := &QYWeiXinReport{
		endpoint:    "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?",
		accessToken: token,
		limiter:     rate.NewLimiter(rate.Every(time.Second*10), 1),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r QYWeiXinReport) Report(ctx context.Context, err error) error {
	err2 := r.limiter.Wait(ctx)
	if err2 != nil {
		return err2
	}
	return r.SendText(err.Error())
}

func (r QYWeiXinReport) SendText(content string) error {
	return r.sendMessage(map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content": content,
		},
		"agentid": 0,
	})
}

func (r QYWeiXinReport) sendMessage(data map[string]any) error {
	s, _ := json.Marshal(data)
	reader := bytes.NewReader(s)
	resp, err := r.client.Post(r.endpoint+"key="+r.accessToken, "application/json", reader)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	var response struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Join(err, fmt.Errorf("statusCode: %d, status: %s", resp.StatusCode, resp.Status))
	}
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return err
	}
	if response.ErrCode != 0 {
		return fmt.Errorf("qyweixin: error response: [%s]", content)
	}
	return nil
}
