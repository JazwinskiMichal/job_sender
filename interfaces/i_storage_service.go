package interfaces

type IStorageService interface {
	// UploadFile uploads a file to a storage bucket.
	UploadFile(objectName string, data []byte, metadata map[string]string) (string, error)

	// DeleteFiles deletes files with the given prefix name.
	DeleteFiles(prefixName string) error
}
