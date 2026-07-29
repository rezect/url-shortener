package testhelpers

import "github.com/rezect/url-shortener/internal/models"

type MockQueue struct{}

func (q *MockQueue) StartWorkers(n int) {}

func (q *MockQueue) Push(click models.Click) {}

func (q *MockQueue) Stop() {}