package types

// Group holds metadata about a group of contractors.
type Group struct {
	ID      string `firestore:"id"`
	OwnerID string `firestore:"owner_id"`

	Name string `firestore:"name"`
}
