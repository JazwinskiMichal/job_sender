package interfaces

import (
	"job_sender/types"
)

// IContractorsDatabaseService is an interface for a database service that manages contractors.
type IContractorsDatabaseService interface {
	// ListContractors lists all for a group.
	ListContractors(groupID string) ([]*types.Contractor, error)

	// GetContractor gets a contractor by ID.
	GetContractor(id string) (*types.Contractor, error)

	// GetContractorsTimesheet gets a contractor's timesheet by ID.
	GetContractorsTimesheet(id string) (*types.Timesheet, error)

	// AddContractor adds a contractor to a group.
	AddContractor(groupID string, contractor *types.Contractor) error

	// AddContractorsTimesheet adds a timesheet to a contractor.
	AddContractorsTimesheet(groupID string, contractorID string, timesheet *types.Timesheet) error

	// UpdateContractor updates a contractor.
	UpdateContractor(contractor *types.Contractor) error

	// DeleteContractor deletes a contractor.
	DeleteContractor(id string) error
}
