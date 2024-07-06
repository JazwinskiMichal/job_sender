package interfaces

import (
	"job_sender/types"
)

// IGroupsDatabaseService is an interface for a database service that manages groups.
type IGroupsDatabaseService interface {
	// GetGroup gets a group by ID.
	GetGroup(id string) (*types.Group, error)

	// AddGroup adds a group.
	AddGroup(group *types.Group) (*types.Group, error)

	// UpdateGroup updates a group.
	UpdateGroup(group *types.Group) error

	// DeleteGroup deletes a group.
	DeleteGroup(id string) error
}
