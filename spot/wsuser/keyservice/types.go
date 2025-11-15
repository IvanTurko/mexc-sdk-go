package keyservice

type listenKeySingle struct {
	ListenKey string `json:"listenKey"`
}

// ListenKeys represents a list of listen keys.
type ListenKeys struct {
	ListenKey []string `json:"listenKey"`
}
