package valueobjects

type Password struct {
	Value string
}

func NewPassword(value string) (*Password, error) {
	// Add validation logic if needed
	return &Password{Value: value}, nil
}
