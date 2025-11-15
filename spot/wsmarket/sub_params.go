package wsmarket

// UpdateInterval defines the update frequency for WebSocket streams.
type UpdateInterval string

const (
	Update10ms  UpdateInterval = "10ms"
	Update100ms UpdateInterval = "100ms"
)

// DepthSize represents the size of an order book depth snapshot.
type DepthSize uint

const (
	DepthLevel5  DepthSize = 5
	DepthLevel10 DepthSize = 10
	DepthLevel20 DepthSize = 20
)

func (u UpdateInterval) isValid() bool {
	switch u {
	case Update10ms, Update100ms:
		return true
	default:
		return false
	}
}

func (d DepthSize) isValid() bool {
	switch d {
	case DepthLevel5, DepthLevel10, DepthLevel20:
		return true
	default:
		return false
	}
}
