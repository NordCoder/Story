package redis

// TODO: make seveFacts with bull insert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NordCoder/Story/internal/infrastructure"
	"github.com/NordCoder/Story/internal/logger"
	"github.com/NordCoder/Story/services/prefetch/category"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/NordCoder/Story/internal/entity"
)

var _ infrastructure.FactRepository = (*FactRepository)(nil)

// FactRepository реализует хранение фактов в Redis с поддержкой транзакций.
// Он зависит только от redis.Cmdable (TxPipeline или обычный клиент)
// и не знает, в транзакции он или нет: это решает Transactor.

type FactRepository struct {
	client *redis.Client // базовый клиент, обычно *redis.Client
	ttl    time.Duration // TTL для отдельного Fact-кеша

	// Ключевые шаблоны — можно переопределить при инициализации, чтобы не жёстко фиксировать строки
	keyFact          string // "fact:%s"  — Hash/JSON по ID
	keyFeedQueue     string // "feed_queue" — Redis List с ID фактов
	categoryProvider category.Provider
}

// NewFactRepository — гибкий конструктор
func NewFactRepository(client *redis.Client, ttl time.Duration, opts ...Option) *FactRepository {
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

// Save сохраняет факт и одновременно пушит его ID в очередь feed_queue.
// Операция атомарна, если вызывается внутри Transactor.WithTx.
func (r *FactRepository) Save(ctx context.Context, f *entity.Fact) error {
	data, err := json.Marshal(f)
	if err != nil {
		logger.LoggerFromContext(ctx).Error("failed to marshal fact for redis", zap.Error(err))
		return fmt.Errorf("marshal fact: %w", err)
	}

	cli := FromContext(ctx, r.client)
	if pipe, ok := cli.(redis.Pipeliner); ok {
		pipe.Set(ctx, r.factKey(f.ID), data, r.ttl)
		pipe.LPush(ctx, r.keyFeedQueue, string(f.ID))
		// Index by category
		pipe.SAdd(ctx, fmt.Sprintf("category_set:%s", f.Category), string(f.ID))
		return nil
	}

	if err := cli.Set(ctx, r.factKey(f.ID), data, r.ttl).Err(); err != nil {
		return err
	}
	if err := cli.LPush(ctx, r.keyFeedQueue, string(f.ID)).Err(); err != nil {
		return err
	}
	// Index by category
	if err := cli.SAdd(ctx, fmt.Sprintf("category_set:%s", f.Category), string(f.ID)).Err(); err != nil {
		logger.LoggerFromContext(ctx).Error("failed to add fact ID to category set", zap.Error(err))
	}
	return nil
}

// GetByID достаёт факт из Redis по ключу.
func (r *FactRepository) GetByID(ctx context.Context, id entity.FactID) (*entity.Fact, error) {
	cli := FromContext(ctx, r.client)

	cmd := cli.Get(ctx, r.factKey(id))
	if err := cmd.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, entity.ErrFactNotFound
		}
		return nil, err
	}
	var f entity.Fact
	if err := json.Unmarshal([]byte(cmd.Val()), &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// GetByCategory returns up to count facts for the given category.
func (r *FactRepository) GetByCategory(ctx context.Context, category entity.Category, count int) ([]*entity.Fact, error) {
	cli := FromContext(ctx, r.client)

	// Получаем случайные ID из множества
	ids, err := cli.SRandMemberN(ctx, fmt.Sprintf("category_set:%s", category), int64(count)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IDs for category %s: %w", category, err)
	}

	if len(ids) == 0 {
		logger.LoggerFromContext(ctx).Info("Add category to provider: " + string(category))
		err := r.categoryProvider.AddCategory(ctx, category)
		if err != nil {
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
				// просто удаляем «висячий» ID
				//todo: logging and handle error
				cli.SRem(ctx, "category_set:"+string(category), idStr)
				continue
			}
			return nil, err
		}
		facts = append(facts, f)
	}
	return facts, nil
}

func (r *FactRepository) PopRandom(ctx context.Context) (*entity.Fact, error) {
	cli := FromContext(ctx, r.client)

	res, err := cli.BRPop(ctx, 100*time.Millisecond, r.keyFeedQueue).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) { // timeout
			return nil, nil
		}
		return nil, err
	}
	if len(res) != 2 {
		return nil, fmt.Errorf("unexpected BRPOP result: %v", res)
	}
	id := entity.FactID(res[1])
	return r.GetByID(ctx, id)
}

func (r *FactRepository) factKey(id entity.FactID) string {
	return fmt.Sprintf(r.keyFact, id)
}

func (r *FactRepository) CountFacts(ctx context.Context) (int64, error) {
	cli := FromContext(ctx, r.client)
	count, err := cli.LLen(ctx, r.keyFeedQueue).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to count facts: %w", err)
	}
	return count, nil
}

func (r *FactRepository) Ping(ctx context.Context) error {
	cli := FromContext(ctx, r.client)
	return cli.Ping(ctx).Err()
}
