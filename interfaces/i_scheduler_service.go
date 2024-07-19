package interfaces

import "job_sender/types"

type ISchedulerService interface {
	// CreateTimesheetRequestJob creates a new Cloud Scheduler job for requesting timesheets.
	CreateTimesheetRequestJob(groupID string, schedule *types.Schedule) error

	// EditTimesheetRequestJob updates a Cloud Scheduler job for requesting timesheets.
	EditTimesheetRequestJob(groupID string, schedule *types.Schedule) error

	// DeleteTimesheetRequestJob deletes a Cloud Scheduler job for requesting timesheets.
	DeleteTimesheetRequestJob(groupID string) error
}
