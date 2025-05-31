package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/prefetch/category"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/NordCoder/Story/internal/entity"
)

type FactRepository struct {
	client           *redis.Client
	ttl              time.Duration
	keyFact          string // шаблон "fact:%s"
	keyFactSet       string // множество всех ID
	categoryProvider category.Provider
}

func NewFactRepository(
	client *redis.Client,
	ttl time.Duration,
	opts ...Option,
) *FactRepository {
	repo := &FactRepository{
		client:     client,
		ttl:        ttl,
		keyFact:    "fact:%s",
		keyFactSet: "all_fact_ids",
	}
	for _, o := range opts {
		o(repo)
	}
	return repo
}

type Option func(*FactRepository)

func WithKeyFact(pattern string) Option { return func(r *FactRepository) { r.keyFact = pattern } }
func WithKeyFactSet(name string) Option { return func(r *FactRepository) { r.keyFactSet = name } }
func WithCategoryProvider(p category.Provider) Option {
	return func(r *FactRepository) { r.categoryProvider = p }
}

func (r *FactRepository) Save(ctx context.Context, f *entity.Fact) error {
	data, err := json.Marshal(f)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to marshal fact", zap.Error(err))
		return fmt.Errorf("marshal fact: %w", err)
	}

	if err := r.client.Set(ctx, r.factKey(f.ID), data, r.ttl).Err(); err != nil {
		return fmt.Errorf("redis SET: %w", err)
	}

	if err := r.client.SAdd(ctx, r.keyFactSet, string(f.ID)).Err(); err != nil {
		return fmt.Errorf("redis SADD (set of facts): %w", err)
	}

	categoryKey := fmt.Sprintf("category_set:%s", f.Category)
	if err := r.client.SAdd(ctx, categoryKey, string(f.ID)).Err(); err != nil {
		logger.LoggerFromContext(ctx).Error("failed to add fact ID to category set", zap.Error(err))
		return fmt.Errorf("redis SADD (category): %w", err)
	}

	return nil
}

func (r *FactRepository) GetByID(ctx context.Context, id entity.FactID) (*entity.Fact, error) {
	cmd := r.client.Get(ctx, r.factKey(id))
	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, entity.ErrFactNotFound
		}
		return nil, fmt.Errorf("redis GET: %w", err)
	}

	var f entity.Fact
	if err := json.Unmarshal([]byte(cmd.Val()), &f); err != nil {
		return nil, fmt.Errorf("unmarshal fact: %w", err)
	}
	return &f, nil
}

func (r *FactRepository) GetByCategory(ctx context.Context, category entity.Category, count int) ([]*entity.Fact, error) {
	categoryKey := fmt.Sprintf("category_set:%s", category)
	ids, err := r.client.SRandMemberN(ctx, categoryKey, int64(count)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IDs for category %s: %w", category, err)
	}

	if len(ids) == 0 {
		logger.LoggerFromContext(ctx).Info("add category to provider: " + string(category))
		if err := r.categoryProvider.AddCategory(ctx, category); err != nil {
			logger.LoggerFromContext(ctx).Error("failed to add category to provider", zap.Error(err))
			return nil, err
		}
		return nil, entity.ErrCategoryNotFound
	}

	var facts []*entity.Fact
	for _, idStr := range ids {
		f, err := r.GetByID(ctx, entity.FactID(idStr))
		if err != nil {
			if errors.Is(err, entity.ErrFactNotFound) {
				_ = r.client.SRem(ctx, categoryKey, idStr).Err()
				_ = r.client.SRem(ctx, r.keyFactSet, idStr).Err()
				continue
			}
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

// PopRandom выбирает случайный живой факт из множества all_fact_ids
func (r *FactRepository) PopRandom(ctx context.Context) (*entity.Fact, error) {
	ids, err := r.client.SRandMemberN(ctx, r.keyFactSet, 5).Result() // берем пачку
	if err != nil {
		return nil, fmt.Errorf("SRANDMEMBER: %w", err)
	}
	for _, id := range ids {
		fact, err := r.GetByID(ctx, entity.FactID(id))
		if err != nil {
			if errors.Is(err, entity.ErrFactNotFound) {
				_ = r.client.SRem(ctx, r.keyFactSet, id).Err()
				continue
			}
			return nil, err
		}
		return fact, nil
	}
	return nil, nil // живых фактов нет
}

func (r *FactRepository) CountFacts(ctx context.Context) (int64, error) {
	count, err := r.client.SCard(ctx, r.keyFactSet).Result()
	if err != nil {
		return 0, fmt.Errorf("SCARD: %w", err)
	}
	return count, nil
}

func (r *FactRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *FactRepository) factKey(id entity.FactID) string {
	return fmt.Sprintf(r.keyFact, id)
}
