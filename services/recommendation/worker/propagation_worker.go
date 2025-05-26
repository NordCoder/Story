package worker

//todo: we need config here

import (
	"context"
	"sync"

	"github.com/NordCoder/Story/internal/logger"
	entity2 "github.com/NordCoder/Story/services/authorization/entity"

	"go.uber.org/zap"

	"github.com/NordCoder/Story/internal/entity"
	"github.com/NordCoder/Story/internal/infrastructure/wikipedia"

	//"github.com/NordCoder/Story/services/recommendation/config"
	"github.com/NordCoder/Story/services/recommendation/repository"
)

type PropagateTask struct {
	UserID   entity2.UserID
	Category entity.Category
	Depth    int
}

type PropagationWorker struct {
	repo       repository.RecRepository
	wikiClient *wikipedia.Client
	//cfg        config.RecommendationConfig
	tasks chan PropagateTask
}

func NewPropagationWorker(
	repo repository.RecRepository,
	wikiClient *wikipedia.Client,
	//cfg config.RecommendationConfig,
	tasks chan PropagateTask,
) *PropagationWorker {
	return &PropagationWorker{
		repo:       repo,
		wikiClient: wikiClient,
		//cfg:        cfg,
		tasks: tasks,
	}
}

func (w *PropagationWorker) Start(ctx context.Context) {
	var wg sync.WaitGroup
	//wg.Add(w.cfg.WorkerCount)
	wg.Add(10)
	//for i := 0; i < w.cfg.WorkerCount; i++ {
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task := <-w.tasks:
					w.handleTask(ctx, task)
				}
			}
		}()
	}
	wg.Wait()
}

func (w *PropagationWorker) handleTask(ctx context.Context, task PropagateTask) {
	//if task.Depth >= w.cfg.MaxDepth {
	if task.Depth >= 1 {
		return
	}

	//subs, err := w.wikiClient.GetSubcategories(ctx, task.Category, w.cfg.SubcatLimit)
	subs, err := w.wikiClient.GetSubcategories(ctx, task.Category, 10)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to get subcategories", zap.String("category", string(task.Category)), zap.Error(err))
		return
	}

	if len(subs) > 0 {
		//weight := math.Pow(w.cfg.DecayFactor, float64(task.Depth))
		//err := w.repo.BulkAdjust(ctx, task.UserID, subs, int(weight))
		if err := w.repo.BulkAdjust(ctx, task.UserID, subs, 1); err != nil {
			logger.LoggerFromContext(ctx).Error("failed to bulk adjust preferences", zap.Error(err), zap.String("category", string(task.Category)))
		}
	}

	for _, sub := range subs {
		nextTask := PropagateTask{
			UserID:   task.UserID,
			Category: sub,
			Depth:    task.Depth + 1,
		}

		select {
		case w.tasks <- nextTask:
		default:
			logger.LoggerFromContext(ctx).Warn("tasks channel full, skipping enqueue", zap.String("subcategory", string(sub)))
		}
	}
}
