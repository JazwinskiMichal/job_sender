package core

import (
	"context"
	"fmt"

	"job_sender/interfaces"
	"job_sender/types"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// TimesheetsDatabaseService is a service for managing timesheets in a database.
type TimesheetsDatabaseService struct {
	collectionName string
	client         *firestore.Client
}

// Ensure TimesheetsDatabaseService implements ITimesheetsDatabaseService.
var _ interfaces.ITimesheetsDatabaseService = &TimesheetsDatabaseService{}

// NewTimesheetsDatabaseService creates a new TimesheetsDatabaseService.
func NewTimesheetsDatabaseService(firebaseService *FirebaseService) (*TimesheetsDatabaseService, error) {
	ctx := context.Background()
	client, err := firebaseService.app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestoredb: could not get Firestore client: %w", err)
	}

	// Verify that we can communicate and authenticate with the Firestore service.
	err = client.RunTransaction(ctx, func(ctx context.Context, t *firestore.Transaction) error {
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("firestoredb: could not connect: %w", err)
	}

	return &TimesheetsDatabaseService{
		collectionName: "timesheets",
		client:         client,
	}, nil
}

// Close closes the database.
func (db *TimesheetsDatabaseService) Close() error {
	return db.client.Close()
}

// ListTimesheets lists all timesheets for a group.
func (db *TimesheetsDatabaseService) ListTimesheets(groupID string) ([]*types.Timesheet, error) {
	ctx := context.Background()
	iter := db.client.Collection(db.collectionName).Where("group_id", "==", groupID).Documents(ctx)

	var timesheets []*types.Timesheet
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		var timesheet types.Timesheet
		err = doc.DataTo(&timesheet)
		if err != nil {
			return nil, fmt.Errorf("could not convert data to timesheet: %w", err)
		}

		timesheets = append(timesheets, &timesheet)
	}

	return timesheets, nil
}

// GetTimesheet gets a timesheet by ID.
func (db *TimesheetsDatabaseService) GetTimesheet(contractorID string, requestID string) (*types.Timesheet, error) {
	ctx := context.Background()
	iter := db.client.Collection(db.collectionName).Where("contractor_id", "==", contractorID).Where("request_id", "==", requestID).Documents(ctx)
	doc, err := iter.Next()
	if err == iterator.Done {
		// Directly return iterator.Done to indicate no timesheet found
		return nil, iterator.Done
	} else if err != nil {
		return nil, fmt.Errorf("could not get timesheet: %w", err)
	}

	var timesheet types.Timesheet
	err = doc.DataTo(&timesheet)
	if err != nil {
		return nil, fmt.Errorf("could not convert data to timesheet: %w", err)
	}

	return &timesheet, nil
}

// GetTimesheetByID gets a timesheet by ID.
func (db *TimesheetsDatabaseService) GetTimesheetByID(id string) (*types.Timesheet, error) {
	ctx := context.Background()
	doc, err := db.client.Collection(db.collectionName).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get timesheet: %w", err)
	}

	var timesheet types.Timesheet
	err = doc.DataTo(&timesheet)
	if err != nil {
		return nil, fmt.Errorf("could not convert data to timesheet: %w", err)
	}

	return &timesheet, nil
}

// AddTimesheet adds a timesheet.
func (db *TimesheetsDatabaseService) AddTimesheet(timesheet *types.Timesheet) error {
	ctx := context.Background()
	ref := db.client.Collection(db.collectionName).NewDoc()
	timesheetMap := map[string]interface{}{
		"id":            ref.ID,
		"contractor_id": timesheet.ContractorID,
		"request_id":    timesheet.RequestID,

		"storage_url": timesheet.StorageURL,
	}

	_, err := ref.Create(ctx, timesheetMap)
	if err != nil {
		return fmt.Errorf("could not add timesheet: %w", err)
	}

	return nil
}

// UpdateTimesheet updates a timesheet.
func (db *TimesheetsDatabaseService) UpdateTimesheet(timesheet *types.Timesheet) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collectionName).Doc(timesheet.ID).Set(ctx, timesheet)
	if err != nil {
		return fmt.Errorf("could not update timesheet: %w", err)
	}

	return nil
}

// DeleteTimesheet deletes a timesheet.
func (db *TimesheetsDatabaseService) DeleteTimesheet(id string) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collectionName).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("could not delete timesheet: %w", err)
	}

	return nil
}
