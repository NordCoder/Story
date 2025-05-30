package prefetch

import (
	"context"
	"github.com/NordCoder/Story/internal/entity"
	"math/rand"
	"time"

	"github.com/NordCoder/Story/services/prefetch/category"
	"github.com/NordCoder/Story/services/prefetch/config"

	"github.com/NordCoder/Story/internal/infrastructure/redis"
	"github.com/NordCoder/Story/internal/infrastructure/wikipedia"
	"go.uber.org/zap"
)

type Prefetcher interface {
	Run(ctx context.Context) error
}

type prefetcher struct {
	cfg                      *config.PrefetcherConfig
	wikipediaClient          wikipedia.WikiClient
	factRepo                 *redis.FactRepository
	logger                   *zap.Logger
	basicCategoryProvider    category.Provider
	advancedCategoryProvider category.Provider
}

// NewPrefetcher создаёт новый экземпляр префетчера.
func NewPrefetcher(
	cfg *config.PrefetcherConfig,
	wikipediaClient wikipedia.WikiClient,
	factRepo *redis.FactRepository,
	logger *zap.Logger,
	basicCategoryProvider category.Provider,
	advancedCategoryProvider category.Provider,
) Prefetcher {
	return &prefetcher{
		cfg:                      cfg,
		wikipediaClient:          wikipediaClient,
		factRepo:                 factRepo,
		logger:                   logger,
		basicCategoryProvider:    basicCategoryProvider,
		advancedCategoryProvider: advancedCategoryProvider,
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

	var concept entity.Category
	r := rand.Float64()
	if r < 0.7 {
		concept, err = p.advancedCategoryProvider.GetCategory(ctx)
	} else {
		concept, err = p.basicCategoryProvider.GetCategory(ctx)
	}

	if err != nil {
		p.logger.Warn("Failed to get category from recommendations", zap.Error(err))
		concept, err = p.basicCategoryProvider.GetCategory(ctx)
	}

	summaries, err := p.wikipediaClient.GetCategorySummaries(ctx, concept, p.cfg.BatchSize)
	if err != nil {
		p.logger.Error("Failed to fetch summaries from Wikipedia", zap.Error(err))
		return err
	}

	for _, summary := range summaries {
		fact := summary.ToFact(concept) // Конвертация ArticleSummary -> Fact

		if err := p.factRepo.Save(ctx, fact); err != nil {
			p.logger.Warn("Failed to save fact", zap.Error(err))
			continue
		}

		p.logger.Info("Saved fact", zap.String("title", fact.Title))
	}

	return nil
}
