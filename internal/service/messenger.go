package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"

	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
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
	repo     persistence.CertificateRepo
	userRepo persistence.UserRepo
	db       *sql.DB
}

func NewMessengerService(repo persistence.CertificateRepo, userRepo persistence.UserRepo, db *sql.DB) *MessengerService {
	return &MessengerService{repo: repo, userRepo: userRepo, db: db}
}

func (s *MessengerService) validateAndSaveUser(ctx *wasmplugin.EventContext, userID int64) (*entity.CertUser, []entity.CertUserPosition, error) {
	users, err := ctx.GetUsersInfo([]int64{userID})
	if err != nil {
		return nil, nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	if len(users) == 0 {
		return nil, nil, apperrors.New(apperrors.KeyUserNotFound, userID)
	}

	info := users[0]
	if info.TsuAccountsID == "" || !info.TsuLinked {
		return nil, nil, apperrors.New(apperrors.KeyTsuNotLinked)
	}
	if info.IsDeanOffice {
		return nil, nil, apperrors.New(apperrors.KeyAccessDenied)
	}
	if info.IsTeacher {
		return nil, nil, apperrors.New(apperrors.KeyAccessDenied)
	}

	isStudent := false
	for _, pos := range info.Positions {
		if pos.PositionType == "student" {
			isStudent = true
			break
		}
	}
	if !isStudent {
		return nil, nil, apperrors.New(apperrors.KeyAccessDenied)
	}

	user := &entity.CertUser{
		ID:            info.ID,
		FullName:      info.FullName,
		ExternalID:    info.ExternalID,
		TsuAccountsID: info.TsuAccountsID,
		TsuLinked:     info.TsuLinked,
		IsTeacher:     info.IsTeacher,
		IsStudent:     info.IsStudent,
		IsDeanOffice:  info.IsDeanOffice,
	}

	var positions []entity.CertUserPosition
	for _, pos := range info.Positions {
		positions = append(positions, entity.CertUserPosition{
			UserID:          info.ID,
			PositionType:    pos.PositionType,
			Status:          pos.Status,
			NationalityType: pos.NationalityType,
			FundingType:     pos.FundingType,
			EducationForm:   pos.EducationForm,
			FacultyName:     pos.FacultyName,
			DepartmentName:  pos.DepartmentName,
			ProgramName:     pos.ProgramName,
			StreamName:      pos.StreamName,
			GroupCode:       pos.GroupCode,
			GroupName:       pos.GroupName,
		})
	}

	_, err = s.userRepo.SaveOrUpdateUser(user, positions)

	if err != nil {
		ctx.LogError(fmt.Sprintf("Save user failed: %+v", err.Error()))
		return nil, nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	return user, positions, nil
}

func (s *MessengerService) SaveOrder(ctx *wasmplugin.EventContext, req CreateOrderRequest) (int64, error) {
	_, _, err := s.validateAndSaveUser(ctx, req.StudentID)
	if err != nil {
		return 0, err
	}

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
			return apperrors.New(apperrors.KeyOrderAlreadyCancelled, req.OrderID)
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
