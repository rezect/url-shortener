package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rezect/url-shortener/internal/models"
	"github.com/rezect/url-shortener/internal/repository"
)

type Queue struct {
	ch            chan models.Click
	repo          *repository.ClickRepository
	wg            sync.WaitGroup
	batchSize     int
	flushInterval time.Duration
}

func NewQueue(ch chan models.Click, repo *repository.ClickRepository, batchSize int, flushInterval time.Duration) *Queue {
	return &Queue{
		ch:            ch,
		repo:          repo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

func (q *Queue) StartWorkers(n int) {
	for i := range n {
		q.wg.Go(func() {
			fmt.Printf("Starting worker %v...\n", i)
			ticker := time.NewTicker(q.flushInterval)
			clicksArr := make([]models.Click, 0, q.batchSize)
			for {
				select {
				case click, ok := <-q.ch:
					if !ok {
						if len(clicksArr) > 0 {
							if err := q.repo.BatchInsert(context.Background(), clicksArr); err != nil {
								fmt.Printf("Error while inserting batches: %v\n", err)
							}
							clicksArr = make([]models.Click, 0, q.batchSize)
						}
						return
					}
					if len(clicksArr) < q.batchSize {
						clicksArr = append(clicksArr, click)
					} else {
						if err := q.repo.BatchInsert(context.Background(), clicksArr); err != nil {
							fmt.Printf("Error while inserting batches: %v\n", err)
						}
						clicksArr = make([]models.Click, 0, q.batchSize)
					}
				case <-ticker.C:
					if err := q.repo.BatchInsert(context.Background(), clicksArr); err != nil {
						fmt.Printf("Error while inserting batches: %v\n", err)
					}
					clicksArr = make([]models.Click, 0, q.batchSize)
				}
			}
		})
	}
}

func (q *Queue) Push(click models.Click) {
	q.ch <- click
}

func (q *Queue) Stop() {
	close(q.ch)
	q.wg.Wait()
}
