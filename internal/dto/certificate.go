package dto

import (
	"time"

	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

type OrderFilter struct {
	CertStatus      *entity.CertificateStatus
	CertificateType string
	DateFrom        *time.Time
	DateTo          *time.Time
	Limit           int
	Offset          int
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
	User        *UserDetails
	Application *entity.CertificateApplication
	Certificate *entity.CertificateDocument
	Attachments []entity.CertificateAttachment
}
