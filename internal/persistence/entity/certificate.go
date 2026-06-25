package entity

type CertificateApplication struct {
	ID              int64                  `json:"id"`
	StudentID       int64                  `json:"student_id"`
	Status          CertificateStatus      `json:"status"`
	Type            CertificateType        `json:"type"`
	ObtainMethod    ObtainMethod           `json:"obtain_method"`
	Comment         string                 `json:"comment,omitempty"`
	RejectionReason string                 `json:"rejection_reason"`
	CreatedAt       string                 `json:"created_at"`
	FormData        map[string]interface{} `json:"form_data" db:"form_data"`
}

type DocumentMetadata struct {
	ID         int64  `json:"id"`
	FileID     string `json:"file_id,omitempty"`
	FileName   string `json:"file_name,omitempty"`
	StorageURL string `json:"storage_url,omitempty"`
	UploadedAt string `json:"created_at"`
}

type CertificateDocument struct {
	DocumentMetadata
	OrderID int64 `json:"order_id"`
}

type CertificateAttachment struct {
	ID         int64  `json:"id"`
	OrderID    int64  `json:"order_id"`
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	FileType   string `json:"file_type"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploaded_at"`
}
