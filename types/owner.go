package types

// Owner holds metadata about the owner of a group.
type Owner struct {
	ID      string `firestore:"id"`
	GroupID string `firestore:"group_id"`

	Name     string `firestore:"name"`
	Surname  string `firestore:"surname"`
	Email    string `firestore:"email"`
	Phone    string `firestore:"phone"`
	PhotoURL string `firestore:"photo_url"`
}
