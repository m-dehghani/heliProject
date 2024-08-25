package valueobjects

type Username struct {
	Value string
}

func NewUsername(value string) (*Username, error) {
	// Add validation logic if needed
	return &Username{Value: value}, nil
}
