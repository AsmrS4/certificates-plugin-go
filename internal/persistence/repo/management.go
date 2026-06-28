package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/AsmrS4/certificates-plugin-go/internal/dto"
	repository "github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	wasmplugin "github.com/SuperBotForge/sdk/go-sdk"
)

var _ repository.ManagementRepo = (*ManagementImpl)(nil)

type ManagementImpl struct {
	db *sql.DB
}

func (r *ManagementImpl) FindWithUserDetails(ctx *wasmplugin.EventContext, id int64) (*dto.CertificateDetails, error) {
	query := `
		SELECT 
			ca.id, ca.student_id, ca.application_status, ca.certificate_type,
			ca.obtain_method, COALESCE(ca.comment, '') as comment,
			COALESCE(ca.rejection_reason, '') as rejection_reason,
			ca.created_at, ca.form_data,
			COALESCE(cu.full_name, '') as full_name,
			COALESCE(cug.nationality_type, '') as nationality_type,
			COALESCE(cug.faculty_name, '') as faculty_name,
			COALESCE(cug.group_code, '') as group_code,
			COALESCE(cug.funding_type, '') as funding_type,
			COALESCE(cug.education_form, '') as education_form,
			COALESCE(cug.stream_name, '') as stream_name,
			COALESCE(cug.status, '') as position_status
		FROM certificate_applications ca
		LEFT JOIN cert_users cu ON ca.student_id = cu.id
		LEFT JOIN cert_user_positions cug ON cu.id = cug.user_id AND cug.position_type = 'student'
		WHERE ca.id = $1
	`

	var details dto.CertificateDetails
	var formDataJSON sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&details.ID,
		&details.StudentID,
		&details.Status,
		&details.Type,
		&details.ObtainMethod,
		&details.Comment,
		&details.RejectionReason,
		&details.CreatedAt,
		&formDataJSON,
		&details.FullName,
		&details.NationalityType,
		&details.FacultyName,
		&details.GroupCode,
		&details.FundingType,
		&details.EducationForm,
		&details.StreamName,
		&details.PositionStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if formDataJSON.Valid && formDataJSON.String != "" {
		str := formDataJSON.String
		if strings.HasPrefix(str, "{") {
			var formData map[string]interface{}
			if err := json.Unmarshal([]byte(str), &formData); err != nil {
				ctx.LogError(fmt.Sprintf("Failed to unmarshal form_data: %v", err))
				details.FormData = make(map[string]interface{})
			} else {
				details.FormData = formData
			}
		} else if strings.HasPrefix(str, "map[") {
			details.FormData = parseMapString(str)
		} else {
			ctx.LogError(fmt.Sprintf("Unexpected format for form_data: %s", str))
			details.FormData = make(map[string]interface{})
		}
	} else {
		details.FormData = make(map[string]interface{})
	}

	ctx.Log(fmt.Sprintf("DEBUG: details.FormData after processing: %+v", details.FormData))

	attachments, err := r.findAttachmentsByOrderID(id)
	if err != nil {
		return nil, err
	}
	details.Attachments = attachments

	certFile, err := r.findCertificateFileByOrderID(id)
	if err != nil {
		return nil, err
	}
	details.CertificateFile = certFile

	return &details, nil
}

func parseMapString(s string) map[string]interface{} {
	s = strings.TrimPrefix(s, "map[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	var key string
	var value strings.Builder
	inValue := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if !inValue {
			if ch == ':' {
				key = strings.TrimSpace(s[:i])
				inValue = true
				value.Reset()

				j := i + 1
				for j < len(s) && s[j] == ' ' {
					j++
				}

				i = j - 1
			}
		} else {
			if ch == ' ' && i+1 < len(s) {
				next := i + 1

				colonIdx := strings.Index(s[next:], ":")
				if colonIdx != -1 {
					potentialKey := s[next : next+colonIdx]
					if !strings.Contains(potentialKey, " ") {
						result[key] = strings.TrimSpace(value.String())
						inValue = false

						i = next - 1
						continue
					}
				}
			}
			value.WriteByte(ch)
		}
	}

	if inValue {
		result[key] = strings.TrimSpace(value.String())
	}

	return result
}

func (r *ManagementImpl) findAttachmentsByOrderID(orderID int64) ([]dto.CertificateAttachmentView, error) {
	rows, err := r.db.Query(`
		SELECT id, file_id, file_name, mime_type, file_type, uploaded_at
		FROM certificate_attachments
		WHERE order_id = $1
		ORDER BY uploaded_at DESC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []dto.CertificateAttachmentView
	for rows.Next() {
		var att dto.CertificateAttachmentView
		err := rows.Scan(
			&att.ID,
			&att.FileID,
			&att.FileName,
			&att.MIMEType,
			&att.FileType,
			&att.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, rows.Err()
}

func (r *ManagementImpl) findCertificateFileByOrderID(orderID int64) (*dto.CertificateFileView, error) {
	var certFile dto.CertificateFileView
	err := r.db.QueryRow(`
		SELECT id, file_id, file_name, storage_url, uploaded_at
		FROM certificate_documents
		WHERE order_id = $1
	`, orderID).Scan(
		&certFile.ID,
		&certFile.FileID,
		&certFile.FileName,
		&certFile.StorageURL,
		&certFile.UploadedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &certFile, nil
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
