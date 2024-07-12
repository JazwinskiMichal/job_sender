package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"job_sender/interfaces"
	"job_sender/types"
	constants "job_sender/utils/constants"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CloudTasksService struct {
	envVariablesService *EnvVariablesService
}

// Ensure CloudTasksService implements ICloudTasksService.
var _ interfaces.ICloudTasksService = &CloudTasksService{}

// NewCloudTasksService creates a new CloudTasksService.
func NewCloudTasksService(envVariablesService *EnvVariablesService) *CloudTasksService {
	return &CloudTasksService{
		envVariablesService: envVariablesService,
	}
}

// CreateTimesheetAggregatorTask creates a new Cloud Task for aggregating timesheets.
func (s *CloudTasksService) CreateTimesheetAggregatorTask(projectID string, locationID string, queueID string, contractor *types.Contractor, model types.TimesheetAggregation) (bool, error) {
	// Check last aggregation time if it is more than 5 minutes ago
	lastAggregationTime := contractor.LastAggregationTimestamp
	currentTime := time.Now().Unix() // Get the current time once to avoid multiple calls

	// If the last aggregation time is not 0 and the difference between the current time and the last aggregation time is less than 5 minutes, return.
	if lastAggregationTime != 0 && currentTime-lastAggregationTime < 300 {
		return false, nil
	}

	// Build the Task queue path.
	queuePath := "projects/" + projectID + "/locations/" + locationID + "/queues/" + queueID

	// Build the Task name.
	taskName := fmt.Sprintf("timesheet-aggregator-%s-%d", contractor.ID, time.Now().Unix())
	name := queuePath + "/tasks/" + taskName

	// Serialize the payload.
	payload, err := json.Marshal(model)
	if err != nil {
		return false, err
	}

	// Create a new Cloud Tasks client.
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return false, err
	}
	defer client.Close()

	// Build the Task payload.
	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			Name: name,
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        constants.AppUrl + "/timesheets/aggregate",
				},
			},
		},
	}

	// Add the payload to the Task.
	req.Task.GetHttpRequest().Body = payload

	// Send the Task to the Cloud Tasks service.
	_, err = client.CreateTask(ctx, req)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return false, nil
		}

		return false, err
	}

	// Update the contractor's last aggregation timestamp.
	contractor.LastAggregationTimestamp = currentTime

	return true, nil
}
