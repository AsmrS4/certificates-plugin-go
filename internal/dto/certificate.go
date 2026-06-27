package dto

import (
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

type FindRequestsFilter struct {
	FullName        string
	NationalityType string
	FacultyName     string
	GroupCode       string
	Status          string
	Type            string
	Limit           int
	Offset          int
}

type CertificateRequestView struct {
	ID              int64  `json:"id"`
	StudentID       int64  `json:"student_id"`
	Status          string `json:"status"`
	Type            string `json:"type"`
	ObtainMethod    string `json:"obtain_method"`
	CreatedAt       string `json:"created_at"`
	FullName        string `json:"full_name"`
	NationalityType string `json:"nationality_type"`
	FacultyName     string `json:"faculty_name"`
	GroupCode       string `json:"group_code"`
}

type CertificateDetails struct {
	ID              int64                       `json:"id"`
	StudentID       int64                       `json:"student_id"`
	Status          string                      `json:"application_status"`
	Type            string                      `json:"certificate_type"`
	ObtainMethod    string                      `json:"obtain_method"`
	Comment         string                      `json:"comment,omitempty"`
	RejectionReason string                      `json:"rejection_reason,omitempty"`
	CreatedAt       string                      `json:"created_at"`
	FormData        map[string]interface{}      `json:"form_data,omitempty"`
	FullName        string                      `json:"full_name"`
	NationalityType string                      `json:"nationality_type"`
	FacultyName     string                      `json:"faculty_name"`
	GroupCode       string                      `json:"group_code"`
	FundingType     string                      `json:"funding_type"`
	EducationForm   string                      `json:"education_form"`
	StreamName      string                      `json:"stream_name"`
	PositionStatus  string                      `json:"position_status"`
	Attachments     []CertificateAttachmentView `json:"attachments,omitempty"`
	CertificateFile *CertificateFileView        `json:"certificate_file,omitempty"`
}

type CertificateAttachmentResponse struct {
	ID         int64  `json:"id"`
	FileID     string `json:"file_id"`
	UploadedAt string `json:"uploaded_at"`
}

type CertificateAttachmentView struct {
	ID         int64  `json:"id"`
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	MIMEType   string `json:"mime_type"`
	FileType   string `json:"file_type"`
	File_URL   string `json:"file_url,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}

type CertificateFileView struct {
	ID         int64  `json:"id"`
	FileID     string `json:"file_id"`
	FileName   string `json:"file_name"`
	StorageURL string `json:"storage_url"`
	UploadedAt string `json:"uploaded_at"`
}

type UserDetails struct {
	FullName        string
	NationalityType string `msgpack:"nationality_type,omitempty" json:"nationality_type,omitempty"`
	FundingType     string `msgpack:"funding_type,omitempty" json:"funding_type,omitempty"`
	EducationForm   string `msgpack:"education_form,omitempty" json:"education_form,omitempty"`
	FacultyName     string `msgpack:"faculty_name,omitempty" json:"faculty_name,omitempty"`
	DepartmentName  string `msgpack:"department_name,omitempty" json:"department_name,omitempty"`
	ProgramName     string `msgpack:"program_name,omitempty" json:"program_name,omitempty"`
	StreamName      string `msgpack:"stream_name,omitempty" json:"stream_name,omitempty"`
	GroupCode       string `msgpack:"group_code,omitempty" json:"group_code,omitempty"`
	GroupName       string `msgpack:"group_name,omitempty" json:"group_name,omitempty"`
}

type OrderDetails struct {
	Application *entity.CertificateApplication
	FileIDs     []string
}

type File struct {
	ID       string
	Name     string
	MIMEType string
	FileType string
}
