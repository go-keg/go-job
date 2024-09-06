package main

import (
	"context"
	"errors"
	"github.com/eiixy/go-job/report"
	"golang.org/x/time/rate"
	syslog "log"
	"os"
	"time"

	"github.com/eiixy/go-job"
	"github.com/go-kratos/kratos/v2/log"
)

func main() {
	j := job.NewJob(
		log.DefaultLogger,
		job.NewWorker("test", example),
		job.NewWorker(
			"test-with-limiter",
			example,
			job.WithLimiter(rate.NewLimiter(rate.Every(time.Second), 10)),
		),
		job.NewWorker(
			"test-with-report-error",
			example,
			job.WithLimiter(rate.NewLimiter(rate.Every(time.Second), 1)),
			job.WithReport(report.NewQYWeiXinReport(os.Getenv("QY_WECHAT_TOKEN"))),
		),
	)
	err := j.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func example(ctx context.Context) error {
	syslog.Println("do something...")
	if time.Now().Second()%10 == 1 {
		// test report error
		return errors.New("test err")
	}
	if time.Now().Second()%25 == 1 {
		// test panic
		panic("test panic")
	}
	return nil
}
