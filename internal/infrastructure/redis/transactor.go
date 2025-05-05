package redis

// NOTE: TURNS OUT THAT IT IS USELESS
//import (
//	"context"
//	"fmt"
//	"github.com/NordCoder/Story/internal/infrastructure"
//	"github.com/NordCoder/Story/internal/logger"
//	"github.com/go-redis/redis/v8"
//
//	"go.uber.org/zap"
//)
//
//var _ infrastructure.Transactor = (*transactorImpl)(nil)
//
//type transactorImpl struct {
//	Client *redis.Client
//}
//
//func NewRedisTransactor(client *redis.Client) *transactorImpl {
//	return &transactorImpl{Client: client}
//}
//
//type ctxTxKey struct{}
//
//func extractTx(ctx context.Context) (redis.Pipeliner, bool) {
//	tx, ok := ctx.Value(ctxTxKey{}).(redis.Pipeliner)
//	return tx, ok
//}
//
//func injectTx(ctx context.Context, pipe redis.Pipeliner) context.Context {
//	return context.WithValue(ctx, ctxTxKey{}, pipe)
//}
//
//func (r *transactorImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
//	if _, ok := extractTx(ctx); ok {
//		return fn(ctx)
//	}
//
//	_, err := r.Client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
//		txCtx := injectTx(ctx, pipe)
//		err := fn(txCtx)
//		if err != nil {
//			logger.LoggerFromContext(ctx).Error("transaction function error, DISCARDING", zap.Error(err))
//			return err
//		}
//		return nil
//	})
//	if err != nil {
//		logger.LoggerFromContext(ctx).Error("Redis transaction failed or aborted", zap.Error(err))
//		return fmt.Errorf("redis transaction error: %w", err)
//	}
//	logger.LoggerFromContext(ctx).Debug("Redis transaction committed successfully")
//	return nil
//}
//
//func FromContext(ctx context.Context, base *redis.Client) redis.Cmdable {
//	if tx, ok := extractTx(ctx); ok {
//		return tx
//	}
//	return base
//}
