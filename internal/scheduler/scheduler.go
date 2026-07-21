package scheduler

import (
	"context"
	"log"
	"math/rand"
	"time"
)

type Scheduler struct{}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) RunEvery(ctx context.Context, base time.Duration, jitter time.Duration, job Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(randJitter(base, jitter)):
			if err := job.Run(ctx); err != nil {
				log.Printf("job failed %v", err)
			}
		}
	}
}

func randJitter(duration, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return duration
	}

	offset := time.Duration(rand.Int63n(int64(2*jitter+1))) - jitter
	return duration + offset
}
