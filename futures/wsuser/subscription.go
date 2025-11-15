package wsuser

type matchFunc func(msg *message) (bool, error)

type wsHandler interface {
	acceptEvent(msg *message) bool
	handleEvent(msg *message)
}

type subscriptionSpec interface {
	wsHandler
	id() string
}

type wsRequestPayload struct {
	Method string `json:"method"`
	Param  any    `json:"param,omitempty"`
}
