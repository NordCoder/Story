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

// FactRepository реализует хранение фактов в Redis.
// Все операции выполняются напрямую через r.client.
type FactRepository struct {
	client           *redis.Client
	ttl              time.Duration
	keyFact          string // шаблон "fact:%s"
	keyFeedQueue     string // имя списка, например "feed_queue"
	categoryProvider category.Provider
}

// NewFactRepository конструктор.
func NewFactRepository(
	client *redis.Client,
	ttl time.Duration,
	opts ...Option,
) *FactRepository {
	repo := &FactRepository{
		client:       client,
		ttl:          ttl,
		keyFact:      "fact:%s",
		keyFeedQueue: "feed_queue",
	}
	for _, o := range opts {
		o(repo)
	}
	return repo
}

type Option func(*FactRepository)

func WithKeyFact(pattern string) Option   { return func(r *FactRepository) { r.keyFact = pattern } }
func WithKeyFeedQueue(name string) Option { return func(r *FactRepository) { r.keyFeedQueue = name } }
func WithCategoryProvider(p category.Provider) Option {
	return func(r *FactRepository) { r.categoryProvider = p }
}

// Save сохраняет факт и пушит его ID в очередь.
func (r *FactRepository) Save(ctx context.Context, f *entity.Fact) error {
	data, err := json.Marshal(f)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to marshal fact", zap.Error(err))
		return fmt.Errorf("marshal fact: %w", err)
	}

	// Сохраняем JSON и ставим TTL
	if err := r.client.Set(ctx, r.factKey(f.ID), data, r.ttl).Err(); err != nil {
		return fmt.Errorf("redis SET: %w", err)
	}

	// Пушим ID в начало очереди
	if err := r.client.LPush(ctx, r.keyFeedQueue, string(f.ID)).Err(); err != nil {
		return fmt.Errorf("redis LPUSH: %w", err)
	}

	// Индекс по категории
	categoryKey := fmt.Sprintf("category_set:%s", f.Category)
	if err := r.client.SAdd(ctx, categoryKey, string(f.ID)).Err(); err != nil {
		logger.LoggerFromContext(ctx).Error("failed to add fact ID to category set", zap.Error(err))
		return fmt.Errorf("redis SADD: %w", err)
	}

	return nil
}

// GetByID достаёт факт по ID.
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

// GetByCategory возвращает до count фактов из множества по категории.
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
				// удаляем "висячий" ID
				if errRem := r.client.SRem(ctx, categoryKey, idStr).Err(); errRem != nil {
					logger.LoggerFromContext(ctx).Error("failed to remove stale ID from category set", zap.Error(errRem))
				}
				continue
			}
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

// PopRandom берёт следующий факт из очереди (с блокирующим ожиданием до 100ms).
func (r *FactRepository) PopRandom(ctx context.Context) (*entity.Fact, error) {
	res, err := r.client.BRPop(ctx, 100*time.Millisecond, r.keyFeedQueue).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// таймаут
			return nil, nil
		}
		return nil, fmt.Errorf("redis BRPOP: %w", err)
	}
	if len(res) != 2 {
		return nil, fmt.Errorf("unexpected BRPOP result: %v", res)
	}
	return r.GetByID(ctx, entity.FactID(res[1]))
}

// CountFacts возвращает длину очереди.
func (r *FactRepository) CountFacts(ctx context.Context) (int64, error) {
	count, err := r.client.LLen(ctx, r.keyFeedQueue).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count facts: %w", err)
	}
	return count, nil
}

// Ping проверяет соединение с Redis.
func (r *FactRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *FactRepository) factKey(id entity.FactID) string {
	return fmt.Sprintf(r.keyFact, id)
}
