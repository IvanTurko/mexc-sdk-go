package wsuser

import (
	"sync"
)

type handlerRouter interface {
	Register(h wsHandler)
	Unregister(h wsHandler)
	Route(msg *message)
	Len() int
}

type handlerRouterImp struct {
	mu       sync.RWMutex
	handlers map[wsHandler]struct{}
}

func newHandlerRouter() *handlerRouterImp {
	return &handlerRouterImp{
		handlers: make(map[wsHandler]struct{}),
	}
}

func (r *handlerRouterImp) Register(h wsHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[h] = struct{}{}
}

func (r *handlerRouterImp) Unregister(h wsHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.handlers, h)
}

func (r *handlerRouterImp) Route(msg *message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for h := range r.handlers {
		if h.acceptEvent(msg) {
			h.handleEvent(msg)
		}
	}
}

func (r *handlerRouterImp) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers)
}
