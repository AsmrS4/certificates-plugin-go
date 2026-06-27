package entity

type CertificateApplication struct {
	ID              int64                  `json:"id"`
	StudentID       int64                  `json:"student_id"`
	Status          CertificateStatus      `json:"application_status"`
	Type            string                 `json:"certificate_type"`
	ObtainMethod    ObtainMethod           `json:"obtain_method"`
	Comment         string                 `json:"comment,omitempty"`
	RejectionReason string                 `json:"rejection_reason"`
	CreatedAt       string                 `json:"created_at"`
	FormData        map[string]interface{} `json:"form_data" db:"form_data"`
}

type CertificateAttachment struct {
	ID         int64  `json:"id"`
	OrderID    int64  `json:"order_id"`
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	MIMEType   string `json:"mime_type"`
	FileType   string `json:"file_type"`
	UploadedAt string `json:"uploaded_at"`
}
