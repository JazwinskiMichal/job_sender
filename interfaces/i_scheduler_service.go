package interfaces

type ISchedulerService interface {
	// CreateTimesheetRequestJob creates a new Cloud Scheduler job for requesting timesheets.
	CreateTimesheetRequestJob(groupID string, schedule string) error

	// DeleteTimesheetRequestJob deletes a Cloud Scheduler job for requesting timesheets.
	DeleteTimesheetRequestJob(groupID string) error
}
