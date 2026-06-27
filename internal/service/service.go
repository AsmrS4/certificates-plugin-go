package service

import (
	"database/sql"

	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
)

type ManagementService struct {
	repo persistence.CertificateRepo
	db   *sql.DB
}

func NewManagementService(repo persistence.CertificateRepo, db *sql.DB) *ManagementService {
	return &ManagementService{repo: repo, db: db}
}
