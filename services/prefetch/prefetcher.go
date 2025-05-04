package prefetch

import (
	"context"
	"github.com/NordCoder/Story/services/prefetch/category"
	"github.com/NordCoder/Story/services/prefetch/config"
	"time"

	"github.com/NordCoder/Story/internal/infrastructure/redis"
	"github.com/NordCoder/Story/internal/infrastructure/wikipedia"
	"go.uber.org/zap"
)

// todo refactor prefatcher for work with new category system
// todo design new category provider based on redis query

type Prefetcher interface {
	Run(ctx context.Context) error
}

type prefetcher struct {
	cfg              *config.PrefetcherConfig
	wikipediaClient  *wikipedia.Client
	factRepo         *redis.FactRepository
	logger           *zap.Logger
	categoryProvider category.Provider
}

// NewPrefetcher создаёт новый экземпляр префетчера.
func NewPrefetcher(
	cfg *config.PrefetcherConfig,
	wikipediaClient *wikipedia.Client,
	factRepo *redis.FactRepository,
	logger *zap.Logger,
	categoryProvider category.Provider,
) Prefetcher {
	return &prefetcher{
		cfg:              cfg,
		wikipediaClient:  wikipediaClient,
		factRepo:         factRepo,
		logger:           logger,
		categoryProvider: categoryProvider,
	}
}

// Run запускает префетчер.
func (p *prefetcher) Run(ctx context.Context) error {
	if !p.cfg.Enabled {
		p.logger.Info("Prefetcher is disabled")
		return nil
	}

	p.logger.Info("Prefetcher started")

	if p.cfg.PrefetchOnStart {
		p.logger.Info("Prefetching on startup...")
		if err := p.prefetch(ctx); err != nil {
			p.logger.Error("Initial prefetch failed", zap.Error(err))
		}
	}

	ticker := time.NewTicker(p.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Prefetcher stopping gracefully")
			return ctx.Err()
		case <-ticker.C:
			if err := p.prefetch(ctx); err != nil {
				p.logger.Error("Prefetch error", zap.Error(err))
			}
		}
	}
}

// prefetch делает одну итерацию загрузки фактов.
func (p *prefetcher) prefetch(ctx context.Context) error {
	count, err := p.factRepo.CountFacts(ctx)
	if err != nil {
		p.logger.Error("Failed to count facts", zap.Error(err))
		return err
	}

	if count >= int64(p.cfg.MinFacts) {
		return nil
	}

	selection, err := p.categoryProvider.GetCategory(ctx)
	if err != nil {
		p.logger.Error("Failed to get category", zap.Error(err))
		return err
	}

	summaries, err := p.wikipediaClient.GetCategorySummaries(ctx, selection.Category, p.cfg.BatchSize)
	if err != nil {
		p.logger.Warn("Failed to fetch summaries from Wikipedia", zap.Error(err))
		return err
	}

	for _, summary := range summaries {
		fact := summary.ToFact(selection.Lang) // Конвертация ArticleSummary -> Fact

		if err := p.factRepo.Save(ctx, fact); err != nil {
			p.logger.Warn("Failed to save fact", zap.Error(err))
			continue
		}

		p.logger.Info("Saved fact", zap.String("title", fact.Title))
	}

	return nil
}
