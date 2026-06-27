package persistence

import (
	"database/sql"

	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

type CertificateRepo interface {
	SaveTx(tx *sql.Tx, c *entity.CertificateApplication) (int64, error)
	SaveAttachmentsTx(tx *sql.Tx, orderID int64, attachments []entity.CertificateAttachment) error
	FindByID(id int64) (*entity.CertificateApplication, error)
	FindAllWithStatus(userID int64, st entity.CertificateStatus, tp string) ([]entity.CertificateApplication, error)
	FindAllByStudent(studentID int64) ([]entity.CertificateApplication, error)
	Cancel(id int64) error
	IsOrderPending(id int64) (bool, error)
	IsCancelled(id int64) (bool, error)
}

type ManagementRepo interface {
	Reject(id int64, reason string) (int64, int64, error)
	Prepare(id int64) (int64, int64, error)
	IsPaper(id int64) (bool, error)
	IsRejected(id int64) (bool, error)
	IsPending(id int64) (bool, error)
	IsExists(id int64) (bool, error)
	IsProcessing(id int64) (bool, error)
	FindRequests(filter *dto.FindRequestsFilter) ([]dto.CertificateRequestView, int64, error)
}

type UserRepo interface {
	SaveOrUpdateUser(user *entity.CertUser, positions []entity.CertUserPosition) (bool, error)
}
