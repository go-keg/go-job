package main

import (
	"context"
	"errors"
	syslog "log"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/time/rate"

	"github.com/eiixy/go-job"
	"github.com/eiixy/go-job/report"
)

func main() {
	j := job.NewJob(log.DefaultLogger,
		job.NewWorker("todo", todo,
			job.WithLimiter(rate.NewLimiter(rate.Every(time.Second), 3)),
			job.WithReport(report.NewQYWeiXinReport(os.Getenv("QY_WECHAT_TOKEN"))),
		),
	)
	// ctx, _ := context.WithTimeout(context.Background(), 8*time.Second)
	ctx := context.Background()
	err := j.Start(ctx)
	if err != nil {
		panic(err)
	}
}

func todo(ctx context.Context) error {
	syslog.Println("do something...")
	if time.Now().Second()%10 == 1 {
		// test report error
		return errors.New("test err")
	}
	return nil
}
