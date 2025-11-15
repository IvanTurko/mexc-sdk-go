package wsmarket

import "encoding/json"

type matchFunc func(msg *message) (bool, error)

type wsHandler interface {
	acceptEvent(msg *PushDataV3MarketWrapper) bool
	handleEvent(msg *PushDataV3MarketWrapper)
}

type subscriptionSpec interface {
	wsHandler
	id() string
	params() any
	matches(msg *message) (bool, error)
}

type subscriptionOp string

const (
	subscribe   subscriptionOp = "SUBSCRIPTION"
	unsubscribe subscriptionOp = "UNSUBSCRIPTION"
)

type wsRequestPayload struct {
	ID     *uint64 `json:"id,omitempty"`
	Method string  `json:"method"`
	Params []any   `json:"params,omitempty"`
}

type subscriptionRequest struct {
	id   uint64
	op   subscriptionOp
	spec subscriptionSpec
}

func newSubscriptionRequest(id uint64, op subscriptionOp, spec subscriptionSpec) *subscriptionRequest {
	return &subscriptionRequest{
		id:   id,
		op:   op,
		spec: spec,
	}
}

func (s *subscriptionRequest) ID() uint64 {
	return s.id
}

func (s *subscriptionRequest) Message() ([]byte, error) {
	payload := wsRequestPayload{
		ID:     &s.id,
		Method: string(s.op),
		Params: []any{s.spec.params()},
	}
	return json.Marshal(payload)
}

func (s *subscriptionRequest) MatchFunc() matchFunc {
	return s.spec.matches
}

var _ wsRequest = (*subscriptionRequest)(nil)
