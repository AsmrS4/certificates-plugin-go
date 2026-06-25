package entity

import "fmt"

type CertificateStatus string
type ObtainMethod string

const (
	Pending   CertificateStatus = "Pending"
	Rejected  CertificateStatus = "Rejected"
	Prepare   CertificateStatus = "Prepare"
	Done      CertificateStatus = "Done"
	Cancelled CertificateStatus = "Cancelled"
)

const (
	Electronic ObtainMethod = "Electronic"
	Paper      ObtainMethod = "Paper"
)

func (c ObtainMethod) IsValid() bool {
	switch c {
	case Electronic, Paper:
		return true
	}
	return false
}

func (c CertificateStatus) IsValid() bool {
	switch c {
	case Pending,
		Rejected,
		Prepare,
		Done,
		Cancelled:
		return true
	}
	return false
}

func (c CertificateStatus) String() string {
	return string(c)
}

func (o ObtainMethod) String() string {
	return string(o)
}

func ParseCertificateStatus(s string) (CertificateStatus, error) {
	cs := CertificateStatus(s)
	if cs.IsValid() {
		return cs, nil
	}
	return "", fmt.Errorf("Invalid certificate status: %s", s)
}

func ParseObtainMethod(s string) (ObtainMethod, error) {
	om := ObtainMethod(s)
	if om.IsValid() {
		return om, nil
	}
	return "", fmt.Errorf("Invalid obtain method: %s", s)
}
