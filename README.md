# go job

## Usage

### Install

```shell
go get github.com/eiixy/go-job
```

### Example
[./example](./example/main.go)
```go
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
	return nil
}

```
