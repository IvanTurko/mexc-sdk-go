package wsutil

import (
	"context"
	"fmt"
	"sync"
)

type PromiseErrSource uint8

const (
	FromUnknown PromiseErrSource = iota
	FromServer
	FromContext
)

func (s PromiseErrSource) String() string {
	switch s {
	case FromServer:
		return "server"
	case FromContext:
		return "context"
	default:
		return "unknown"
	}
}

type PromiseError struct {
	Source PromiseErrSource
	Err    error
}

func (e *PromiseError) Error() string {
	return fmt.Sprintf("promise failed [%s]: %v", e.Source, e.Err)
}

func (e *PromiseError) Unwrap() error {
	return e.Err
}

type Promise[T any] interface {
	Match(msg *T) (bool, error)
	Resolve(msg *T)
	Reject(err error)
	Await(ctx context.Context) (*T, error)
}

func NewPromise[T any](matchFn func(*T) (bool, error)) Promise[T] {
	return &promiseImp[T]{
		matchFn: matchFn,
		resCh:   make(chan *T, 1),
		errCh:   make(chan error, 1),
	}
}

type promiseImp[T any] struct {
	matchFn func(*T) (bool, error)
	resCh   chan *T
	errCh   chan error
	once    sync.Once
}

func (p *promiseImp[T]) Match(msg *T) (bool, error) {
	return p.matchFn(msg)
}

func (p *promiseImp[T]) Resolve(msg *T) {
	p.once.Do(func() {
		p.resCh <- msg
	})
}

func (p *promiseImp[T]) Reject(err error) {
	p.once.Do(func() {
		p.errCh <- err
	})
}

func (p *promiseImp[T]) Await(ctx context.Context) (*T, error) {
	select {
	case msg := <-p.resCh:
		return msg, nil
	case err := <-p.errCh:
		return nil, &PromiseError{Source: FromServer, Err: err}
	case <-ctx.Done():
		return nil, &PromiseError{Source: FromContext, Err: ctx.Err()}
	}
}
