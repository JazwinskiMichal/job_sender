package core

import (
	"context"
	"fmt"
	"job_sender/interfaces"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type StorageService struct {
	storageBucketName string
	storageBucket     *storage.BucketHandle
}

// Ensure StorageService implements the IStorageService interface.
var _ interfaces.IStorageService = &StorageService{}

// NewStorageService creates a new StorageService.
func NewStorageService(bucketName string) (*StorageService, error) {
	ctx := context.Background()

	// Create a new Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	// Get a handle for the bucket
	bucket := client.Bucket(bucketName)

	// Check if the bucket exists
	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			return nil, fmt.Errorf("bucket %s does not exist", bucketName)
		}
		return nil, fmt.Errorf("could not get bucket %s: %v", bucketName, err)
	}

	return &StorageService{
		storageBucketName: bucketName,
		storageBucket:     bucket,
	}, nil
}

// UploadFile uploads a file to a storage bucket.
func (s *StorageService) UploadFile(objectName string, data []byte, metadata map[string]string) (string, error) {
	ctx := context.Background()

	// Create a new object in the bucket
	obj := s.storageBucket.Object(objectName)
	w := obj.NewWriter(ctx)

	// Set custom metadata
	if metadata != nil {
		w.ObjectAttrs.Metadata = metadata
	}

	// TODO: storage.AllUsers gives public read access to anyone.
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = "application/octet-stream"

	// Entries are immutable, be aggressive about caching (1 day).
	w.CacheControl = "public, max-age=86400"

	// Write the data to the object
	_, err := w.Write(data)
	if err != nil {
		return "", fmt.Errorf("could not write data to object: %v", err)
	}

	// Close the writer
	err = w.Close()
	if err != nil {
		return "", fmt.Errorf("could not close writer: %v", err)
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, s.storageBucketName, objectName), nil
}

// DeleteFiles deletes files with the given prefix name.
func (s *StorageService) DeleteFiles(prefixName string) error {
	ctx := context.Background()

	// Get all the objects in the bucket
	it := s.storageBucket.Objects(ctx, &storage.Query{Prefix: prefixName})

	// Delete all the objects with the given prefix
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("could not get next object: %v", err)
		}

		err = s.storageBucket.Object(objAttrs.Name).Delete(ctx)
		if err != nil {
			return fmt.Errorf("could not delete object: %v", err)
		}
	}

	return nil
}
