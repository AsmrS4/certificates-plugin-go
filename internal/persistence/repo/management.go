package repo

import (
	"database/sql"
	"fmt"

	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	repository "github.com/AsmrS4/certificates-plugin-go/internal/persistence"
)

var _ repository.ManagementRepo = (*ManagementImpl)(nil)

type ManagementImpl struct {
	db *sql.DB
}

func NewManagementRepo(db *sql.DB) *ManagementImpl {
	return &ManagementImpl{db: db}
}

func (r *ManagementImpl) FindRequests(filter *dto.FindRequestsFilter) ([]dto.CertificateRequestView, int64, error) {
	query := `
        SELECT 
            ca.id, ca.student_id, ca.application_status, ca.certificate_type, 
            ca.obtain_method, ca.created_at,
            COALESCE(cu.full_name, '') as full_name,
            COALESCE(cug.nationality_type, '') as nationality_type,
            COALESCE(cug.faculty_name, '') as faculty_name,
            COALESCE(cug.group_code, '') as group_code
        FROM certificate_applications ca
        LEFT JOIN cert_users cu ON ca.student_id = cu.id
        LEFT JOIN cert_user_positions cug ON cu.id = cug.user_id AND cug.position_type = 'student'
        WHERE ca.application_status != 'Cancelled'
    `
	countQuery := `
        SELECT COUNT(*)
        FROM certificate_applications ca
        LEFT JOIN cert_users cu ON ca.student_id = cu.id
        LEFT JOIN cert_user_positions cug ON cu.id = cug.user_id AND cug.position_type = 'student'
        WHERE ca.application_status != 'Cancelled'
    `
	args := []interface{}{}
	argIdx := 1

	if filter.FullName != "" {
		query += fmt.Sprintf(" AND cu.full_name ILIKE $%d", argIdx)
		countQuery += fmt.Sprintf(" AND cu.full_name ILIKE $%d", argIdx)
		args = append(args, "%"+filter.FullName+"%")
		argIdx++
	}
	if filter.NationalityType != "" {
		query += fmt.Sprintf(" AND cug.nationality_type = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND cug.nationality_type = $%d", argIdx)
		args = append(args, filter.NationalityType)
		argIdx++
	}
	if filter.FacultyName != "" {
		query += fmt.Sprintf(" AND cug.faculty_name = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND cug.faculty_name = $%d", argIdx)
		args = append(args, filter.FacultyName)
		argIdx++
	}
	if filter.GroupCode != "" {
		query += fmt.Sprintf(" AND cug.group_code = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND cug.group_code = $%d", argIdx)
		args = append(args, filter.GroupCode)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND ca.application_status = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND ca.application_status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Type != "" {
		query += fmt.Sprintf(" AND ca.certificate_type = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND ca.certificate_type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query += fmt.Sprintf(" ORDER BY ca.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []dto.CertificateRequestView
	for rows.Next() {
		var view dto.CertificateRequestView
		err := rows.Scan(
			&view.ID,
			&view.StudentID,
			&view.Status,
			&view.Type,
			&view.ObtainMethod,
			&view.CreatedAt,
			&view.FullName,
			&view.NationalityType,
			&view.FacultyName,
			&view.GroupCode,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, view)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int64
	err = r.db.QueryRow(countQuery, args[:argIdx-1]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
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
