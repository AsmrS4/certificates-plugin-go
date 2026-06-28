// internal/service/management.go
package service

import (
	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
)

type ManagementService struct {
	managementRepo persistence.ManagementRepo
	userRepo       *persistence.UserRepo
}

func NewManagementService(managementRepo persistence.ManagementRepo, userRepo *persistence.UserRepo) *ManagementService {
	return &ManagementService{
		managementRepo: managementRepo,
		userRepo:       userRepo,
	}
}

func (s *ManagementService) FindRequests(filter dto.FindRequestsFilter) ([]dto.CertificateRequestView, int64, error) {
	results, total, err := s.managementRepo.FindRequests(&filter)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (s *ManagementService) FindWithUserDetails(ctx *wasmplugin.EventContext, id int64) (*dto.CertificateDetails, error) {
	if id <= 0 {
		return nil, apperrors.New(apperrors.KeyInvalidID, id)
	}

	details, err := s.managementRepo.FindWithUserDetails(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	if details == nil {
		return nil, apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	return details, nil
}

func (s *ManagementService) RejectOrder(id int64, reason string) (int64, error) {
	exists, err := s.managementRepo.IsExists(id)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	_, err = s.managementRepo.IsRejected(id)
	if err != nil {
		return 0, err
	}

	pending, err := s.managementRepo.IsPending(id)
	if err != nil {
		return 0, err
	}
	if !pending {
		return 0, apperrors.New(apperrors.KeyOrderNotPending, id)
	}
	_, studentID, err := s.managementRepo.Reject(id, reason)
	if err != nil {
		return 0, err
	}
	return studentID, nil
}

func (s *ManagementService) PrepareOrder(id int64) (int64, error) {
	exists, err := s.managementRepo.IsExists(id)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	pending, err := s.managementRepo.IsPending(id)
	if err != nil {
		return 0, err
	}
	if !pending {
		return 0, apperrors.New(apperrors.KeyOrderNotPending, id)
	}
	_, studentID, err := s.managementRepo.Prepare(id)
	if err != nil {
		return 0, err
	}
	return studentID, nil
}

func (s *ManagementService) UploadCertificateFile(orderID int64, file dto.File) (int64, error) {
	exists, err := s.managementRepo.IsExists(orderID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, apperrors.New(apperrors.KeyOrderNotFound, orderID)
	}
	processing, err := s.managementRepo.IsProcessing(orderID)
	if err != nil {
		return 0, err
	}
	if !processing {
		return 0, apperrors.New(apperrors.KeyOrderNotInProcessing, orderID)
	}
	//метод для сохранения файла
	//
	return 1, nil
}
