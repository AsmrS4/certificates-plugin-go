package repo

import (
	"database/sql"

	repository "github.com/AsmrS4/certificates-plugin-go/internal/persistence"
)

var _ repository.ManagementRepo = (*ManagementImpl)(nil)

type ManagementImpl struct {
	db *sql.DB
}

func (r *ManagementImpl) Prepare(id int64) (int64, int64, error) {
	var orderID, studentID int64
	err := r.db.QueryRow(`
        UPDATE certificate_applications
        SET application_status = 'Prepare'
        WHERE id = $1
        RETURNING id, student_id
    `, id).Scan(&orderID, &studentID)
	if err != nil {
		return 0, 0, err
	}
	return orderID, studentID, nil
}

func (r *ManagementImpl) Reject(id int64, reason string) (int64, int64, error) {
	var orderID, studentID int64
	err := r.db.QueryRow(`
        UPDATE certificate_applications
        SET application_status = 'Rejected', rejection_reason = $1
        WHERE id = $2
        RETURNING id, student_id
    `, reason, id).Scan(&orderID, &studentID)
	if err != nil {
		return 0, 0, err
	}
	return orderID, studentID, nil
}

func (r *ManagementImpl) IsExists(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1
        )`, id).Scan(&exists)
	return exists, err
}

func (r *ManagementImpl) IsPaper(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND obtain_method = 'Paper'
        )`, id).Scan(&exists)
	return exists, err
}

func (r *ManagementImpl) IsPending(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND application_status = 'Pending'
        )`, id).Scan(&exists)
	return exists, err
}

func (r *ManagementImpl) IsRejected(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND application_status = 'Rejected'
        )`, id).Scan(&exists)
	return exists, err
}

func (r *ManagementImpl) IsProcessing(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND application_status = 'Prepare'
        )`, id).Scan(&exists)
	return exists, err
}
