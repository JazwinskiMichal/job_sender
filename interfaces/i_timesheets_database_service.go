package interfaces

import (
	"job_sender/types"
)

// ITimesheetsDatabaseService is an interface for a database service that manages timesheets.
type ITimesheetsDatabaseService interface {
	// ListTimesheets lists all timesheets for a group.
	ListTimesheets(groupID string) ([]*types.Timesheet, error)

	// GetTimesheet gets a timesheet by ContractorID and RequestID.
	GetTimesheet(contractorID string, requestID string) (*types.Timesheet, error)

	// GetTimesheetByID gets a timesheet by ID.
	GetTimesheetByID(id string) (*types.Timesheet, error)

	// AddTimesheet adds a timesheet to a group.
	AddTimesheet(timesheet *types.Timesheet) error

	// UpdateTimesheet updates a timesheet.
	UpdateTimesheet(timesheet *types.Timesheet) error

	// DeleteTimesheet deletes a timesheet.
	DeleteTimesheet(id string) error
}
