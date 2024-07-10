package interfaces

type IStorageService interface {
	// UploadFile uploads a file to a storage bucket.
	UploadFile(objectName string, data []byte) (string, error)
}
