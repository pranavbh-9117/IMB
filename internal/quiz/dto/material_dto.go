package dto

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// MaterialResponse represents the metadata of an uploaded quiz material.
type MaterialResponse struct {
	ID               uuid.UUID `json:"id"`
	QuizID           uuid.UUID `json:"quiz_id"`
	UploadedBy       uuid.UUID `json:"uploaded_by"`
	OriginalFilename string    `json:"original_filename"`
	StoredFilename   string    `json:"stored_filename"`
	ContentType      string    `json:"content_type"`
	FileSize         int64     `json:"file_size"`
	CreatedAt        time.Time `json:"created_at"`
}

// UploadMaterialsResponse represents the response after successfully uploading materials.
type UploadMaterialsResponse struct {
	Uploaded []MaterialResponse `json:"uploaded"`
}

// DownloadResult represents the stream and headers needed to download a file.
type DownloadResult struct {
	Reader      io.ReadCloser
	Filename    string
	ContentType string
	Size        int64
}
