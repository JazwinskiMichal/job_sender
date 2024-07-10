package types

// Contractor holds metadata about a contractor.
type Contractor struct {
	ID      string `firestore:"id"`
	GroupID string `firestore:"group_id"`

	Name     string `firestore:"name"`
	Surname  string `firestore:"surname"`
	Email    string `firestore:"email"`
	Phone    string `firestore:"phone"`
	PhotoURL string `firestore:"photo_url"`

	LastAggregationTimestamp int64 `firestore:"last_aggregation_timestamp"`
}
