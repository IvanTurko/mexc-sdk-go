package wsuser

import "fmt"

// PositionType represents the type of a position.
type PositionType string

const (
	PositionTypeLong  PositionType = "LONG"
	PositionTypeShort PositionType = "SHORT"
)

func parsePositionType(code int) (PositionType, error) {
	switch code {
	case 1:
		return PositionTypeLong, nil
	case 2:
		return PositionTypeShort, nil
	default:
		return "", fmt.Errorf("unknown position type code: %d", code)
	}
}

// OpenType represents the open type of a position.
type OpenType string

const (
	OpenTypeIsolated OpenType = "ISOLATED"
	OpenTypeCross    OpenType = "CROSS"
)

func parseOpenType(code int) (OpenType, error) {
	switch code {
	case 1:
		return OpenTypeIsolated, nil
	case 2:
		return OpenTypeCross, nil
	default:
		return "", fmt.Errorf("unknown open type code: %d", code)
	}
}
