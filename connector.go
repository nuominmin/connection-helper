package connectionhelper

import (
	"context"
)

// Connector .
type Connector[T any] interface {
	Connect() (T, error)              // 建立连接
	IsConnectionError(err error) bool // 判断是否为连接错误
}

// Operation 定义了要执行的操作，它接受一个泛型连接并返回错误
type Operation[T any] func(ctx context.Context, conn T) error

// RetryExecutor 提供了重试逻辑
type RetryExecutor[T any] struct {
	conn       T            // 连接
	retryLimit int          // 重试限制
	Connector  Connector[T] // 使用泛型的Connector接口
}

// New 创建一个新的RetryExecutor实例
func New[T any](connector Connector[T]) (*RetryExecutor[T], error) {
	conn, err := connector.Connect()
	if err != nil {
		return nil, err
	}

	return &RetryExecutor[T]{
		retryLimit: 3,
		Connector:  connector,
		conn:       conn,
	}, nil
}

// SetRetryLimit 设置重试次数（默认是3次）
func (r *RetryExecutor[T]) SetRetryLimit(limit int) {
	if limit <= 0 {
		limit = 3
	}
	r.retryLimit = limit
}

// ExecWithRetry 使用提供的Connector重试操作
func (r *RetryExecutor[T]) ExecWithRetry(ctx context.Context, op Operation[T]) error {
	var err error

	for i := 0; i < r.retryLimit; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err = op(ctx, r.conn); err == nil {
			return nil
		}

		if !r.Connector.IsConnectionError(err) {
			return err
		}

		if r.conn, err = r.Connector.Connect(); err != nil {
			return err
		}
	}
	return err // 返回最后一次的错误
}

// Exec 使用提供的Connector操作
func (r *RetryExecutor[T]) Exec(ctx context.Context, op Operation[T]) error {
	return op(ctx, r.conn)
}

// Close .
func (r *RetryExecutor[T]) Close(close func(conn T) error) error {
	return close(r.conn)
}
