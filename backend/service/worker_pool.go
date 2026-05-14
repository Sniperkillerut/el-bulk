package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/el-bulk/backend/models"
	"github.com/el-bulk/backend/utils/logger"
)

type JobHandler func(ctx context.Context, job *models.Job, updateProgress func(int)) (models.JSONB, error)

type WorkerPool struct {
	JobService *JobService
	handlers   map[string]JobHandler
	queue      chan *models.Job
	wg         sync.WaitGroup
	numWorkers int
	stopChan   chan struct{}
}

func NewWorkerPool(js *JobService, numWorkers int) *WorkerPool {
	return &WorkerPool{
		JobService: js,
		handlers:   make(map[string]JobHandler),
		queue:      make(chan *models.Job, 100),
		numWorkers: numWorkers,
		stopChan:   make(chan struct{}),
	}
}

func (p *WorkerPool) RegisterHandler(jobType string, handler JobHandler) {
	p.handlers[jobType] = handler
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *WorkerPool) Stop() {
	close(p.stopChan)
	p.wg.Wait()
}

func (p *WorkerPool) Submit(job *models.Job) {
	p.queue <- job
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	logger.Info("[worker-%d] started", id)

	for {
		select {
		case job := <-p.queue:
			p.processJob(id, job)
		case <-p.stopChan:
			logger.Info("[worker-%d] stopping", id)
			return
		}
	}
}

func (p *WorkerPool) processJob(workerID int, job *models.Job) {
	handler, ok := p.handlers[job.Type]
	if !ok {
		err := fmt.Errorf("no handler registered for job type: %s", job.Type)
		logger.Error("[worker-%d] %v", workerID, err)
		p.JobService.UpdateStatus(context.Background(), job.ID, models.JobFailed, 0, nil, err)
		return
	}

	logger.Info("[worker-%d] processing job %s (type: %s)", workerID, job.ID, job.Type)
	p.JobService.UpdateStatus(context.Background(), job.ID, models.JobRunning, 0, nil, nil)

	ctx := context.Background() // In a real app, we might want to attach tracing here
	
	updateProgress := func(progress int) {
		p.JobService.Store.UpdateProgress(ctx, job.ID, progress)
	}

	result, err := handler(ctx, job, updateProgress)
	if err != nil {
		logger.Error("[worker-%d] job %s failed: %v", workerID, job.ID, err)
		p.JobService.UpdateStatus(ctx, job.ID, models.JobFailed, 0, nil, err)
		return
	}

	logger.Info("[worker-%d] job %s completed", workerID, job.ID)
	p.JobService.UpdateStatus(ctx, job.ID, models.JobCompleted, 100, result, nil)
}
