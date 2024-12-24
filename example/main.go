package main

import (
	"context"
	"errors"
	syslog "log"
	"os"
	"time"

	job "github.com/go-keg/go-job"
	"github.com/go-keg/go-job/report/qyweixin"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/time/rate"
)

func main() {
	var index int
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
			job.WithReport(qyweixin.NewReport(os.Getenv("QY_WECHAT_TOKEN"))),
		),
		job.NewWorker("test-loop", func(ctx context.Context) error {
			limiter := rate.NewLimiter(rate.Every(2*time.Second), 3)
			index++
			for {
				select {
				case <-ctx.Done():
					return nil
				default:
					if limiter.Allow() {
						if time.Now().Second()%30 == 0 {
							return errors.New("test-loop error")
						}
						syslog.Println(index, "do...")
					} else {
						time.Sleep(time.Second)
					}
				}
			}
		}, job.WithLimiterDuration(10*time.Second)),
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
