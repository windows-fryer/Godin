package service

var Services = []string{
	"iris",
}

type CDNBackend interface {
	// UploadFile uploads a file to the CDN backend.
	UploadFile(sessionToken string, data []byte) error

	// DownloadFile downloads a file from the CDN backend using the part ID.
	DownloadFile(partID string) ([]byte, error)
}
