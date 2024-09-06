package job

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Job struct {
	works []*Worker
	log   *log.Helper
}

func NewJob(logger log.Logger, works ...*Worker) *Job {
	return &Job{
		log:   log.NewHelper(log.With(logger, "module", "job")),
		works: works,
	}
}

func (j Job) Start(ctx context.Context) error {
	if len(j.works) == 0 {
		return nil
	}
	var wg sync.WaitGroup
	for _, item := range j.works {
		wg.Add(1)
		go func(ctx2 context.Context, worker *Worker) {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				default:
					if worker.reStartLimiter.Allow() {
						if os.Getenv("JOB_ENABLE") != "false" {
							j.log.Info("start job:", worker.name)
							j.run(ctx2, worker)
						}
					} else {
						time.Sleep(worker.sleep)
					}
				}
			}
		}(ctx, item)
	}
	wg.Wait()
	return nil
}

func (j Job) run(ctx context.Context, work *Worker) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			j.log.Error(err, "panic", "stack", "...\n"+string(buf))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if work.limiter.Allow() {
				err := work.job(ctx)
				if err != nil {
					if work.reportError != nil {
						reportErr := work.reportError.Report(ctx, err)
						if reportErr != nil {
							j.log.Errorw("method", "report", "err", reportErr)
						}
					}
					j.log.Error(err)
				}
			} else {
				time.Sleep(work.sleep)
			}
		}
	}
}
