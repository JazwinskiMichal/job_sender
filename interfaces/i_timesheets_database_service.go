package interfaces

import (
	"job_sender/types"
)

// ITimesheetsDatabaseService is an interface for a database service that manages timesheets.
type ITimesheetsDatabaseService interface {
	// ListTimesheets lists all timesheets for a group.
	ListTimesheets(groupID string) ([]*types.Timesheet, error)

	// GetTimesheet gets a timesheet by ID.
	GetTimesheet(id string) (*types.Timesheet, error)

	// AddTimesheet adds a timesheet to a group.
	AddTimesheet(timesheet *types.Timesheet) error

	// UpdateTimesheet updates a timesheet.
	UpdateTimesheet(timesheet *types.Timesheet) error

	// DeleteTimesheet deletes a timesheet.
	DeleteTimesheet(id string) error
}
