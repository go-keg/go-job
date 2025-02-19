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
	works       []*Worker
	reportError ReportError
	log         *log.Helper
	cancel      context.CancelFunc
}

func NewJob(logger log.Logger, works ...*Worker) *Job {
	return &Job{
		log:   log.NewHelper(log.With(logger, "module", "job")),
		works: works,
	}
}

func NewJobWithReport(logger log.Logger, report ReportError, works ...*Worker) *Job {
	return &Job{
		log:         log.NewHelper(log.With(logger, "module", "job")),
		works:       works,
		reportError: report,
	}
}

func (j *Job) Start(ctx context.Context) error {
	if len(j.works) == 0 {
		return nil
	}
	ctx, j.cancel = context.WithCancel(ctx)
	var wg sync.WaitGroup
	for _, item := range j.works {
		wg.Add(1)
		go func(worker *Worker) {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				default:
					if worker.reStartLimiter.Allow() {
						if os.Getenv("JOB_ENABLE") != "false" {
							j.log.Debug("start job:", worker.name)
							j.run(ctx, worker)
						}
					} else {
						time.Sleep(worker.sleep)
					}
				}
			}
		}(item)
	}
	wg.Wait()
	return nil
}

func (j *Job) Stop() {
	if j.cancel != nil {
		j.cancel()
	}
}

func (j *Job) run(ctx context.Context, work *Worker) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			j.log.Errorw("worker", work.name, "errType", "panic", "error", err, "stack", string(buf))
			report := work.reportError
			if report == nil && j.reportError != nil {
				report = j.reportError
			}
			if report != nil {
				reportErr := report.Report(ctx, work.name, fmt.Sprintf("panic:\nerror: %s\nstack: %s", err, buf))
				if reportErr != nil {
					j.log.Errorw("method", "report", "worker", work.name, "err", reportErr)
				}
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if work.limiter.Allow() {
				j.log.Debugw("method", "run_worker", "worker", work.name)
				err := work.job(ctx)
				if err != nil {
					report := work.reportError
					if report == nil && j.reportError != nil {
						report = j.reportError
					}
					if report != nil {
						reportErr := report.Report(ctx, work.name, err.Error())
						if reportErr != nil {
							j.log.Errorw("method", "report", "worker", work.name, "err", reportErr)
						}
					}
					j.log.Errorw("worker", work.name, "error", err)
				}
			} else {
				time.Sleep(work.sleep)
			}
		}
	}
}
