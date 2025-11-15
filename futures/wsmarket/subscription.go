package wsmarket

import (
	"encoding/json"
)

type matchFunc func(msg *message) (bool, error)

type wsHandler interface {
	acceptEvent(msg *message) bool
	handleEvent(msg *message)
}

type subscriptionSpec interface {
	wsHandler
	matches(msg *message) (bool, error)
	payload(op subscriptionOp) any
	id() string
}

type subscriptionOp string

const (
	subscribe   subscriptionOp = "sub"
	unsubscribe subscriptionOp = "unsub"
)

type wsRequestPayload struct {
	Method string `json:"method"`
	Param  any    `json:"param,omitempty"`
	Gzip   bool   `json:"gzip,omitempty"`
}

type subscriptionRequest struct {
	op   subscriptionOp
	spec subscriptionSpec
}

func newSubscriptionRequest(op subscriptionOp, spec subscriptionSpec) *subscriptionRequest {
	return &subscriptionRequest{
		op:   op,
		spec: spec,
	}
}

func (s *subscriptionRequest) Message() ([]byte, error) {
	return json.Marshal(s.spec.payload(s.op))
}

func (s *subscriptionRequest) MatchFunc() matchFunc {
	return s.spec.matches
}

var _ wsRequest = (*subscriptionRequest)(nil)
