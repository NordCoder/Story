package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NordCoder/Story/internal/entity"
	"github.com/go-redis/redis/v8"
)

// FactRepository хранит факты в Redis без транзакций.
type FactRepository struct {
	client         *redis.Client
	ttl            time.Duration
	keyFactPattern string // шаблон ключа для факта, например "fact:%s"
	keyFeedQueue   string // имя списка-очереди, например "feed_queue"
}

// NewFactRepository создаёт репозиторий фактов без транзакций.
func NewFactRepository(
	client *redis.Client,
	ttl time.Duration,
	opts ...Option,
) *FactRepository {
	r := &FactRepository{
		client:         client,
		ttl:            ttl,
		keyFactPattern: "fact:%s",
		keyFeedQueue:   "feed_queue",
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Option для настройки репозитория.
type Option func(*FactRepository)

// WithKeyFact переопределяет шаблон ключа факта.
func WithKeyFact(pattern string) Option {
	return func(r *FactRepository) {
		r.keyFactPattern = pattern
	}
}

// WithKeyFeedQueue задаёт имя очереди.
func WithKeyFeedQueue(name string) Option {
	return func(r *FactRepository) {
		r.keyFeedQueue = name
	}
}

// Save сохраняет факт и пушит его ID в очередь.
func (r *FactRepository) Save(ctx context.Context, f *entity.Fact) error {
	data, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal fact: %w", err)
	}
	key := fmt.Sprintf(r.keyFactPattern, f.ID)

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	if err := r.client.LPush(ctx, r.keyFeedQueue, string(f.ID)).Err(); err != nil {
		return fmt.Errorf("redis LPUSH %s: %w", r.keyFeedQueue, err)
	}
	return nil
}

var ErrNotFound = errors.New("not found")
var ErrQueueEmpty = errors.New("no facts in queue")

// PopRandom FactRepository.PopRandom
func (r *FactRepository) PopRandom(ctx context.Context) (*entity.Fact, error) {
	res, err := r.client.RPop(ctx, r.keyFeedQueue).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrQueueEmpty
		}
		return nil, fmt.Errorf("redis RPOP %s: %w", r.keyFeedQueue, err)
	}

	fact, err := r.GetByID(ctx, entity.FactID(res))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return fact, nil
}

// GetByID Обновлённый GetByID для единообразия
func (r *FactRepository) GetByID(ctx context.Context, id entity.FactID) (*entity.Fact, error) {
	key := fmt.Sprintf(r.keyFactPattern, id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("redis GET %s: %w", key, err)
	}
	var f entity.Fact
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("unmarshal fact %s: %w", key, err)
	}
	return &f, nil
}

// CountFacts возвращает длину очереди.
func (r *FactRepository) CountFacts(ctx context.Context) (int64, error) {
	n, err := r.client.LLen(ctx, r.keyFeedQueue).Result()
	if err != nil {
		return 0, fmt.Errorf("redis LLEN %s: %w", r.keyFeedQueue, err)
	}
	return n, nil
}

// Ping проверяет соединение с Redis.
func (r *FactRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
