package core

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"job_sender/interfaces"
	"job_sender/types"
)

type ContractorsDatabaseService struct {
	collection string
	client     *firestore.Client
}

// Ensure ContractorsDatabaseService implements IContractorsDatabaseService.
var _ interfaces.IContractorsDatabaseService = &ContractorsDatabaseService{}

// NewContractorsDatabaseService creates a new ContractorsDatabaseService.
func NewContractorsDatabaseService(firebaseService *FirebaseService) (*ContractorsDatabaseService, error) {
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

	return &ContractorsDatabaseService{
		collection: "contractors",
		client:     client,
	}, nil
}

// Close closes the database.
func (db *ContractorsDatabaseService) Close(context.Context) error {
	return db.client.Close()
}

// ListContractors lists all contractors for a group.
func (db *ContractorsDatabaseService) ListContractors(groupID string) ([]*types.Contractor, error) {
	ctx := context.Background()
	iter := db.client.Collection(db.collection).Where("group_id", "==", groupID).Documents(ctx)

	var contractors []*types.Contractor
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("firestoredb: could not list contractors: %w", err)
		}

		contractor := &types.Contractor{}
		if err := doc.DataTo(contractor); err != nil {
			return nil, fmt.Errorf("firestoredb: could not convert data to contractor: %w", err)
		}

		contractors = append(contractors, contractor)
	}

	return contractors, nil
}

// GetContractor gets a contractor by ID.
func (db *ContractorsDatabaseService) GetContractor(id string) (*types.Contractor, error) {
	ctx := context.Background()
	doc, err := db.client.Collection(db.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestoredb: could not get contractor: %w", err)
	}

	contractor := &types.Contractor{}
	if err := doc.DataTo(contractor); err != nil {
		return nil, fmt.Errorf("firestoredb: could not convert data to contractor: %w", err)
	}

	return contractor, nil
}

// AddContractor adds a contractor to a group.
func (db *ContractorsDatabaseService) AddContractor(groupID string, contractor *types.Contractor) error {
	ctx := context.Background()

	// Check if contractor already exists.
	iter := db.client.Collection(db.collection).Where("email", "==", contractor.Email).Documents(ctx)
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("firestoredb: could not check if contractor exists: %w", err)
		}

		return fmt.Errorf("firestoredb: contractor already exists")
	}

	ref := db.client.Collection(db.collection).NewDoc()
	contractorMap := map[string]interface{}{
		"id":        ref.ID,
		"role":      contractor.Role,
		"group_id":  groupID,
		"name":      contractor.Name,
		"surname":   contractor.Surname,
		"email":     contractor.Email,
		"phone":     contractor.Phone,
		"photo_url": contractor.PhotoURL,
	}

	_, err := ref.Create(ctx, contractorMap)
	if err != nil {
		return fmt.Errorf("firestoredb: could not add contractor: %w", err)
	}

	return nil
}

// UpdateContractor updates a contractor.
func (db *ContractorsDatabaseService) UpdateContractor(contractor *types.Contractor) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collection).Doc(contractor.ID).Set(ctx, map[string]interface{}{
		"role":      contractor.Role,
		"group_id":  contractor.GroupID,
		"name":      contractor.Name,
		"surname":   contractor.Surname,
		"email":     contractor.Email,
		"phone":     contractor.Phone,
		"photo_url": contractor.PhotoURL,
	})
	if err != nil {
		return fmt.Errorf("firestoredb: could not update contractor: %w", err)
	}

	return nil
}

// DeleteContractor deletes a contractor.
func (db *ContractorsDatabaseService) DeleteContractor(id string) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collection).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestoredb: could not delete contractor: %w", err)
	}

	return nil
}
