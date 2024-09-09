package qyweixin

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

type Report struct {
	endpoint    string
	accessToken string
	client      *http.Client
	limiter     *rate.Limiter
}

type ReportFunc func(report *Report)

func WithLimiter(limiter *rate.Limiter) ReportFunc {
	return func(report *Report) {
		report.limiter = limiter
	}
}

func WithTimeout(duration time.Duration) ReportFunc {
	return func(report *Report) {
		report.client.Timeout = duration
	}
}

func NewReport(token string, opts ...ReportFunc) *Report {
	r := &Report{
		endpoint:    "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?",
		accessToken: token,
		limiter:     rate.NewLimiter(rate.Every(time.Second*10), 1),
		client:      &http.Client{Timeout: time.Second * 3},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r Report) Report(ctx context.Context, err error) error {
	return r.sendText(ctx, err.Error())
}

func (r Report) sendText(ctx context.Context, content string) error {
	return r.sendMessage(ctx, map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content": content,
		},
		"agentid": 0,
	})
}

func (r Report) sendMessage(ctx context.Context, data map[string]any) error {
	err := r.limiter.Wait(ctx)
	if err != nil {
		return err
	}
	s, _ := json.Marshal(data)
	reader := bytes.NewReader(s)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint+"key="+r.accessToken, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := r.client.Do(request)
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
