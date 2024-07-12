package core

import (
	"context"
	"fmt"

	"job_sender/interfaces"
	"job_sender/types"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupsDatabaseService struct {
	collectionName string
	client         *firestore.Client
}

// Ensure GroupsDatabaseService implements IGroupsDatabaseService.
var _ interfaces.IGroupsDatabaseService = &GroupsDatabaseService{}

// NewGroupsDatabaseService creates a new GroupsDatabaseService.
func NewGroupsDatabaseService(firebaseService *FirebaseService) (*GroupsDatabaseService, error) {
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

	return &GroupsDatabaseService{
		collectionName: "groups",
		client:         client,
	}, nil
}

// Close closes the database.
func (db *GroupsDatabaseService) Close(context.Context) error {
	return db.client.Close()
}

// GetGroup gets a group by ID.
func (db *GroupsDatabaseService) GetGroup(id string) (*types.Group, error) {
	ctx := context.Background()
	doc, err := db.client.Collection(db.collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Directly return the NotFound error
			return nil, status.Errorf(codes.NotFound, "group with ID %s does not exist", id)
		}
		return nil, fmt.Errorf("could not get group: %w", err)
	}

	var group types.Group
	err = doc.DataTo(&group)
	if err != nil {
		return nil, fmt.Errorf("could not convert group data: %w", err)
	}

	return &group, nil
}

// AddGroup adds a group.
func (db *GroupsDatabaseService) AddGroup(group *types.Group) (*types.Group, error) {
	ctx := context.Background()

	// Check if the group already exists.
	if group.ID != "" {
		_, err := db.client.Collection(db.collectionName).Doc(group.ID).Get(ctx)
		if err != nil {
			if status.Code(err) == codes.AlreadyExists {
				return nil, status.Errorf(codes.AlreadyExists, "group with ID %s already exists", group.ID)
			} else if status.Code(err) == codes.NotFound {
				// Continue
			} else {
				return nil, fmt.Errorf("could not get group: %w", err)
			}
		}
	}

	ref := db.client.Collection(db.collectionName).NewDoc()
	groupMap := map[string]interface{}{
		"id":       ref.ID,
		"name":     group.Name,
		"owner_id": group.OwnerID,
	}

	_, err := ref.Create(ctx, groupMap)
	if err != nil {
		return nil, fmt.Errorf("could not add group: %w", err)
	}

	// Update the group with the ID.
	group.ID = ref.ID

	return group, nil
}

// UpdateGroup updates a group.
func (db *GroupsDatabaseService) UpdateGroup(group *types.Group) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collectionName).Doc(group.ID).Set(ctx, group)
	if err != nil {
		return fmt.Errorf("could not update group: %w", err)
	}

	return nil
}

// DeleteGroup deletes a group and all contractors.
func (db *GroupsDatabaseService) DeleteGroup(id string) error {
	ctx := context.Background()

	// Get the group.
	doc, err := db.client.Collection(db.collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Errorf(codes.NotFound, "group with ID %s does not exist", id)
		}
		return fmt.Errorf("could not get group: %w", err)
	}

	var group types.Group
	err = doc.DataTo(&group)
	if err != nil {
		return fmt.Errorf("could not convert group data: %w", err)
	}

	// Delete all contractors in the group.
	contractors, err := db.client.Collection("contractors").Where("group_id", "==", group.ID).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("could not get contractors: %w", err)
	}

	for _, contractor := range contractors {
		contractorID := contractor.Ref.ID

		_, err := db.client.Collection("contractors").Doc(contractorID).Delete(ctx)
		if err != nil {
			return fmt.Errorf("could not delete contractor: %w", err)
		}
	}

	// Delete the group.
	_, err = db.client.Collection(db.collectionName).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("could not delete group: %w", err)
	}

	return nil
}
