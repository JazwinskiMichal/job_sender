package types

// Contractor holds metadata about a contractor.
type Contractor struct {
	ID      string
	Role    string
	GroupID string

	Name     string
	Surname  string
	Email    string
	Phone    string
	PhotoURL string
}
