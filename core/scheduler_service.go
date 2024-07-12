package core

import (
	"context"
	"fmt"

	"job_sender/interfaces"
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
func (s *SchedulerService) CreateTimesheetRequestJob(groupID string, schedule string) error {
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
		Schedule: schedule, //"0 17 * * SUN",
	}

	// Create the job
	req := &schedulerpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", s.projectID, s.location),
		Job:    job,
	}

	_, err := s.client.CreateJob(context.Background(), req)
	if err != nil {
		return fmt.Errorf("CreateJob: %v", err)
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
