package core

import (
	"context"
	"fmt"

	"job_sender/interfaces"
	"job_sender/types"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OwnerDatabaseService struct {
	collectionName string
	client         *firestore.Client
}

// Ensure OwnerDatabaseService implements IOwnerDatabaseService.
var _ interfaces.IOwnerDatabaseService = &OwnerDatabaseService{}

// NewOwnerDatabaseService creates a new OwnerDatabaseService.
func NewOwnerDatabaseService(firebaseService *FirebaseService) (*OwnerDatabaseService, error) {
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

	return &OwnerDatabaseService{
		collectionName: "owners",
		client:         client,
	}, nil
}

// Close closes the database.
func (db *OwnerDatabaseService) Close(context.Context) error {
	return db.client.Close()
}

// GetOwnerByEmail gets an owner by email.
func (db *OwnerDatabaseService) GetOwnerByEmail(email string) (*types.Owner, error) {
	ctx := context.Background()
	iter := db.client.Collection(db.collectionName).Where("email", "==", email).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return nil, status.Errorf(codes.NotFound, "owner with email %s does not exist", email)
			}
			return nil, fmt.Errorf("firestoredb: could not get owner: %w", err)
		}

		// Get the first owner with the specified email.
		owner := &types.Owner{}
		if err := doc.DataTo(owner); err != nil {
			return nil, fmt.Errorf("firestoredb: could not convert data to owner: %w", err)
		}

		return owner, nil
	}

	return nil, status.Errorf(codes.NotFound, "owner with email %s does not exist", email)
}

// GetOwnerByID gets an owner by ID.
func (db *OwnerDatabaseService) GetOwnerByID(id string) (*types.Owner, error) {
	ctx := context.Background()
	doc, err := db.client.Collection(db.collectionName).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, status.Errorf(codes.NotFound, "owner with ID %s does not exist", id)
		}
		return nil, fmt.Errorf("firestoredb: could not get owner: %w", err)
	}

	var owner types.Owner
	if err := doc.DataTo(&owner); err != nil {
		return nil, fmt.Errorf("firestoredb: could not convert owner data: %w", err)
	}

	return &owner, nil
}

// AddOwner adds an owner.
func (db *OwnerDatabaseService) AddOwner(owner *types.Owner) error {
	ctx := context.Background()

	// Reference to the document with the specified owner ID.
	docRef := db.client.Collection(db.collectionName).Doc(owner.ID)

	// Check if the owner already exists.
	_, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// If not found, proceed to create the new owner document with the specified ID.
			_, err = docRef.Set(ctx, map[string]interface{}{
				"id":        owner.ID,
				"name":      owner.Name,
				"surname":   owner.Surname,
				"email":     owner.Email,
				"phone":     owner.Phone,
				"photo_url": owner.PhotoURL,
			})
			if err != nil {
				return fmt.Errorf("firestoredb: could not add owner: %w", err)
			}
		} else {
			// Handle other errors.
			return fmt.Errorf("firestoredb: could not get owner: %w", err)
		}
	} else {
		// Owner already exists.
		return status.Errorf(codes.AlreadyExists, "owner with ID %s already exists", owner.ID)
	}

	return nil
}

// UpdateOwner updates an owner.
func (db *OwnerDatabaseService) UpdateOwner(owner *types.Owner) error {
	ctx := context.Background()
	_, err := db.client.Collection(db.collectionName).Doc(owner.ID).Set(ctx, owner)
	if err != nil {
		return fmt.Errorf("firestoredb: could not update owner: %w", err)
	}

	return nil
}

// DeleteOwner deletes an owner, group and all contractors.
func (db *OwnerDatabaseService) DeleteOwner(id string) error {
	ctx := context.Background()

	// Get the owner's groups.
	groups, err := db.client.Collection("groups").Where("owner_id", "==", id).Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("firestoredb: could not get groups: %w", err)
	}

	// Delete the owner's groups and contractors.
	for _, group := range groups {
		groupID := group.Ref.ID

		// Get the group's contractors.
		contractors, err := db.client.Collection("contractors").Where("group_id", "==", groupID).Documents(ctx).GetAll()
		if err != nil {
			return fmt.Errorf("firestoredb: could not get contractors: %w", err)
		}

		// Delete the group's contractors.
		for _, contractor := range contractors {
			contractorID := contractor.Ref.ID

			_, err := db.client.Collection("contractors").Doc(contractorID).Delete(ctx)
			if err != nil {
				return fmt.Errorf("firestoredb: could not delete contractor: %w", err)
			}
		}

		// Delete the group.
		_, err = db.client.Collection("groups").Doc(groupID).Delete(ctx)
		if err != nil {
			return fmt.Errorf("firestoredb: could not delete group: %w", err)
		}
	}

	// Delete the owner.
	_, err = db.client.Collection(db.collectionName).Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("firestoredb: could not delete owner: %w", err)
	}

	return nil
}
