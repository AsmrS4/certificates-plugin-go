// internal/service/management.go
package service

import (
	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	apperrors "github.com/AsmrS4/certificates-plugin-go/internal/errors"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence"
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

func (s *ManagementService) FindWithUserDetails(id int64) (*dto.CertificateDetails, error) {
	if id <= 0 {
		return nil, apperrors.New(apperrors.KeyInvalidID, id)
	}

	details, err := s.managementRepo.FindWithUserDetails(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.KeyInternalError, err)
	}
	if details == nil {
		return nil, apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	return details, nil
}

func (s *ManagementService) RejectOrder(id int64, reason string) error {
	exists, err := s.managementRepo.IsExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	_, err = s.managementRepo.IsRejected(id)
	if err != nil {
		return err
	}

	pending, err := s.managementRepo.IsPending(id)
	if err != nil {
		return err
	}
	if !pending {
		return apperrors.New(apperrors.KeyOrderNotPending, id)
	}
	_, _, err = s.managementRepo.Reject(id, reason)
	if err != nil {
		return err
	}
	return nil
}

func (s *ManagementService) PrepareOrder(id int64) error {
	exists, err := s.managementRepo.IsExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	pending, err := s.managementRepo.IsPending(id)
	if err != nil {
		return err
	}
	if !pending {
		return apperrors.New(apperrors.KeyOrderNotPending, id)
	}
	_, _, err = s.managementRepo.Prepare(id)
	if err != nil {
		return err
	}
	return nil
}

func (s *ManagementService) FinishPaperOrder(id int64) error {
	exists, err := s.managementRepo.IsExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.New(apperrors.KeyOrderNotFound, id)
	}
	processing, err := s.managementRepo.IsProcessing(id)
	if err != nil {
		return err
	}
	if !processing {
		return apperrors.New(apperrors.KeyOrderNotInProcessing, id)
	}
	_, _, err = s.managementRepo.Prepare(id)
	if err != nil {
		return err
	}
	return nil
}
