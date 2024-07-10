package interfaces

import (
	"job_sender/types"
)

// ICloudTasksService is an interface for a Cloud Tasks service.
type ICloudTasksService interface {
	// CreateTimesheetAggregatorTask creates a new Cloud Task for aggregating timesheets.
	CreateTimesheetAggregatorTask(projectID string, locationID string, queueID string, contractor *types.Contractor, payload types.TimesheetAggregation) (bool, error)
}
