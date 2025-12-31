package filequeue

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrQueueFull   = errors.New("filequeue: queue is full")
	ErrQueueClosed = errors.New("filequeue: queue is closed")
)

type Job struct {
	ID  string
	Run func(ctx context.Context)
}

type Queue struct {
	jobs    chan Job
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	once    sync.Once
}

func NewQueue(cfg Config) *Queue {
	if cfg.Workers <= 0 {
		cfg.Workers = defaultWorkers
	}
	if cfg.Size <= 0 {
		cfg.Size = defaultSize
	}
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		jobs:    make(chan Job, cfg.Size),
		workers: cfg.Workers,
		ctx:     ctx,
		cancel:  cancel,
	}
	q.start()
	return q
}

func (q *Queue) Enqueue(job Job) error {
	if job.Run == nil {
		return errors.New("filequeue: job is nil")
	}
	select {
	case <-q.ctx.Done():
		return ErrQueueClosed
	default:
	}
	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Close() {
	if q == nil {
		return
	}
	q.once.Do(func() {
		q.cancel()
		close(q.jobs)
		q.wg.Wait()
	})
}

func (q *Queue) start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Go(func() {
			for {
				select {
				case <-q.ctx.Done():
					return
				case job, ok := <-q.jobs:
					if !ok {
						return
					}
					if job.Run != nil {
						job.Run(q.ctx)
					}
				}
			}
		})
	}
}
