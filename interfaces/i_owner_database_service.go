package interfaces

import (
	"job_sender/types"
)

// IOwnerDatabaseService is an interface for a database service that manages owners.
type IOwnerDatabaseService interface {
	// GetOwnerByEmail gets an owner by email.
	GetOwnerByEmail(email string) (*types.Owner, error)

	// GetOwner gets an owner by ID.
	GetOwnerByID(id string) (*types.Owner, error)

	// AddOwner adds an owner.
	AddOwner(owner *types.Owner) error

	// UpdateOwner updates an owner.
	UpdateOwner(owner *types.Owner) error

	// DeleteOwner deletes an owner.
	DeleteOwner(id string) error
}
