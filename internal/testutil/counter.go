package testutil

type MockCounter struct {
	Value    uint64
	IncValue uint64
}

func (m *MockCounter) Get() uint64 {
	return m.Value
}

func (m *MockCounter) Inc() uint64 {
	return m.IncValue
}
