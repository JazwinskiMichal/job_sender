package core

import (
	"context"
	"fmt"
	"strings"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"

	scheduler "cloud.google.com/go/scheduler/apiv1"
	"cloud.google.com/go/scheduler/apiv1/schedulerpb"
	"google.golang.org/api/option"
)

type SchedulerService struct {
	serviceAccountEmail string
	projectID           string
	location            string

	client *scheduler.CloudSchedulerClient
}

// Ensure SchedulerService implements ISchedulerService
var _ interfaces.ISchedulerService = &SchedulerService{}

// NewSchedulerService creates a new SchedulerService
func NewSchedulerService(serviceAccountEmail string, projectID string, location string, secretServiceAccountKey []byte) (*SchedulerService, error) {
	ctx := context.Background()

	// Create credentials from the service account key
	cred := option.WithCredentialsJSON(secretServiceAccountKey)

	// Create a Cloud Scheduler client with credentials
	client, err := scheduler.NewCloudSchedulerClient(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("scheduler.NewCloudSchedulerClient: %v", err)
	}

	return &SchedulerService{
		serviceAccountEmail: serviceAccountEmail,
		projectID:           projectID,
		location:            location,

		client: client,
	}, nil
}

// CreateTimesheetRequestJob creates a new Cloud Scheduler job for requesting timesheets.
func (s *SchedulerService) CreateTimesheetRequestJob(groupID string, schedule *types.Schedule) error {
	// Convert the schedule to a cron expression
	cronExpression, err := convertScheduleToCron(schedule)
	if err != nil {
		return fmt.Errorf("convertScheduleToCron: %v", err)
	}

	// Define the job to be created
	job := &schedulerpb.Job{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.projectID, s.location, fmt.Sprintf("timesheet-request-scheduler-job-%s", groupID)),
		Target: &schedulerpb.Job_HttpTarget{
			HttpTarget: &schedulerpb.HttpTarget{
				Uri:        constants.AppUrl + "/timesheets/request?groupID=" + groupID,
				HttpMethod: schedulerpb.HttpMethod_POST,
				AuthorizationHeader: &schedulerpb.HttpTarget_OidcToken{
					OidcToken: &schedulerpb.OidcToken{
						ServiceAccountEmail: s.serviceAccountEmail,
						Audience:            constants.AppUrl + "/timesheets/request",
					},
				},
			},
		},
		Schedule: cronExpression,
		TimeZone: schedule.Timezone,
	}

	// Create the job
	req := &schedulerpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", s.projectID, s.location),
		Job:    job,
	}

	_, err = s.client.CreateJob(context.Background(), req)
	if err != nil {
		return fmt.Errorf("CreateJob: %v", err)
	}

	return nil
}

// EditTimesheetRequestJob updates a Cloud Scheduler job for requesting timesheets.
func (s *SchedulerService) EditTimesheetRequestJob(groupID string, schedule *types.Schedule) error {
	// Generate the job name
	jobName := fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.projectID, s.location, fmt.Sprintf("timesheet-request-scheduler-job-%s", groupID))

	// Retrieve the existing job
	job, err := s.client.GetJob(context.Background(), &schedulerpb.GetJobRequest{Name: jobName})
	if err != nil {
		return fmt.Errorf("GetJob: %v", err)
	}

	// Convert the new schedule to a cron expression
	cronExpression, err := convertScheduleToCron(schedule)
	if err != nil {
		return fmt.Errorf("convertScheduleToCron: %v", err)
	}

	// Update the job's schedule
	job.Schedule = cronExpression

	// Update the job's timezone
	job.TimeZone = schedule.Timezone

	// Update the job
	_, err = s.client.UpdateJob(context.Background(), &schedulerpb.UpdateJobRequest{Job: job})
	if err != nil {
		return fmt.Errorf("UpdateJob: %v", err)
	}

	return nil
}

// DeleteTimesheetRequestJob deletes a Cloud Scheduler job for requesting timesheets.
func (s *SchedulerService) DeleteTimesheetRequestJob(groupID string) error {
	// Delete the job
	req := &schedulerpb.DeleteJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", s.projectID, s.location, fmt.Sprintf("timesheet-request-scheduler-job-%s", groupID)),
	}

	err := s.client.DeleteJob(context.Background(), req)
	if err != nil {
		return fmt.Errorf("DeleteJob: %v", err)
	}

	return nil
}

// convertScheduleToCron translates a Schedule instance to a Unix-cron format string.
// Adjusted to handle intervals greater than 1 elsewhere.
func convertScheduleToCron(s *types.Schedule) (string, error) {
	// Split the Time into hour and minute
	timeParts := strings.Split(s.Time, ":")
	if len(timeParts) != 2 {
		return "", fmt.Errorf("invalid time format")
	}
	hour, minute := timeParts[0], timeParts[1]

	// Default values for cron fields
	dayOfMonth, month, dayOfWeek := "*", "*", "*"

	switch s.IntervalType {
	case constants.Weeks:
		// Assuming Weekday is already validated and converted elsewhere
		dayOfWeek = weekdayToCronDay(s.Weekday)
	case constants.Months:
		// Assuming Monthday is already validated elsewhere
		dayOfMonth = s.Monthday
	}

	cronExpression := fmt.Sprintf("%s %s %s %s %s", minute, hour, dayOfMonth, month, dayOfWeek)

	return cronExpression, nil
}

// weekdayToCronDay converts a weekday string to a cron-compatible day-of-week.
func weekdayToCronDay(weekday string) string {
	switch weekday {
	case "Sunday":
		return "0"
	case "Monday":
		return "1"
	case "Tuesday":
		return "2"
	case "Wednesday":
		return "3"
	case "Thursday":
		return "4"
	case "Friday":
		return "5"
	case "Saturday":
		return "6"
	default:
		return ""
	}
}
