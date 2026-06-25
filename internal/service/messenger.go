package service

import (
	"database/sql"
	"errors"
	"time"

	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

type CreateOrderRequest struct {
	StudentID       int64
	CertificateType string
	ObtainMethod    string
	Comment         string
	FormData        map[string]interface{}
	Files           []FileInfo
}

type FileInfo struct {
	ID       string
	Name     string
	FileType string
	MimeType string
	Size     int64
}

type RejectOrderRequest struct {
	OrderID int64
}

type FindOrderRequest struct {
	OrderID int64
}

type FindAllRequest struct {
	StudentID int64
	Status    string
	Type      string
}

type MessengerService struct {
	repo persistence.CertificateRepo
	db   *sql.DB
}

func NewMessengerService(repo persistence.CertificateRepo, db *sql.DB) *MessengerService {
	return &MessengerService{repo: repo, db: db}
}

func (s *MessengerService) SaveOrder(req CreateOrderRequest) (int64, error) {
	order := &entity.CertificateApplication{
		StudentID:    req.StudentID,
		Type:         req.CertificateType,
		ObtainMethod: entity.ObtainMethod(req.ObtainMethod),
		Status:       entity.Pending,
		Comment:      req.Comment,
		FormData:     req.FormData,
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	orderID, err := s.repo.SaveTx(tx, order)
	if err != nil {
		return 0, apperrors.Wrap(apperrors.KeyInternalError, err)
	}

	if len(req.Files) > 0 {
		attachments := make([]entity.CertificateAttachment, len(req.Files))
		now := time.Now().Format(time.RFC3339)
		for i, f := range req.Files {
			attachments[i] = entity.CertificateAttachment{
				OrderID:    orderID,
				FileID:     f.ID,
				FileName:   f.Name,
				FileType:   f.FileType,
				MimeType:   f.MimeType,
				Size:       f.Size,
				UploadedAt: now,
			}
		}
		if err := s.repo.SaveAttachmentsTx(tx, orderID, attachments); err != nil {
			return 0, apperrors.Wrap(apperrors.KeyAttachmentFailed, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, apperrors.Wrap(apperrors.KeyInternalError, err)
	}

	return orderID, nil
}

func (s *MessengerService) RejectOrder(req RejectOrderRequest) error {
	pending, err := s.repo.IsOrderPending(req.OrderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.New(apperrors.KeyOrderNotFound, req.OrderID)
		}
		return apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	cancelled, err := s.repo.IsCancelled(req.OrderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.New(apperrors.KeyOrderNotFound, req.OrderID)
		}
		return apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	if !pending {
		if cancelled {
			apperrors.New(apperrors.KeyOrderAlreadyCancelled, req.OrderID)
		}
		return apperrors.New(apperrors.KeyOrderNotPending, req.OrderID)
	}
	if err := s.repo.Cancel(req.OrderID); err != nil {
		return apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	return nil
}

func (s *MessengerService) FindOrder(req FindOrderRequest) (*entity.CertificateApplication, error) {
	order, err := s.repo.FindByID(req.OrderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.New(apperrors.KeyOrderNotFound, req.OrderID)
		}
		return nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	if order == nil {
		return nil, apperrors.New(apperrors.KeyOrderNotFound, req.OrderID)
	}
	return order, nil
}

func (s *MessengerService) FindAll(req FindAllRequest) ([]*entity.CertificateApplication, error) {
	var orders []entity.CertificateApplication
	var err error

	if req.Status != "" && req.Type != "" && req.Status != "Skip" {
		st := entity.CertificateStatus(req.Status)
		orders, err = s.repo.FindAllWithStatus(req.StudentID, st, req.Type)
	} else {
		orders, err = s.repo.FindAllByStudent(req.StudentID)
	}
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}

	result := make([]*entity.CertificateApplication, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}
	return result, nil
}
