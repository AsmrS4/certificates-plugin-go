package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"

	repository "github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

var _ repository.CertificateRepo = (*CertificateImpl)(nil)

type CertificateImpl struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *CertificateImpl {
	return &CertificateImpl{db: db}
}

func (r *CertificateImpl) SaveTx(tx *sql.Tx, c *entity.CertificateApplication) (int64, error) {
	if c.FormData == nil {
		c.FormData = make(map[string]interface{})
	}
	formDataJSON, err := json.Marshal(c.FormData)
	if err != nil {
		return 0, fmt.Errorf("marshal form_data: %w", err)
	}

	var id int64
	err = tx.QueryRow(
		`INSERT INTO certificate_applications(
            student_id, certificate_type, obtain_method, comment, form_data
        ) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		c.StudentID, c.Type, c.ObtainMethod, c.Comment, formDataJSON,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert application: %w", err)
	}
	return id, nil
}

func (r *CertificateImpl) SaveAttachmentsTx(tx *sql.Tx, orderID int64, attachments []entity.CertificateAttachment) error {
	if len(attachments) == 0 {
		return nil
	}
	for _, att := range attachments {
		_, err := tx.Exec(
			`INSERT INTO certificate_attachments(order_id, file_id, file_name, mime_type, file_type) VALUES ($1, $2, $3, $4, $5)`,
			orderID, att.FileID, att.FileName, att.MIMEType, att.FileType,
		)
		if err != nil {
			return fmt.Errorf("insert attachment: %w", err)
		}
	}
	return nil
}

func (c *CertificateImpl) Cancel(id int64) error {
	_, err := c.db.Exec(`UPDATE certificate_applications SET application_status = 'Cancelled' WHERE id = $1`, id)
	return err
}

func (r *CertificateImpl) IsOrderPending(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND application_status = 'Pending'
        )`, id).Scan(&exists)
	return exists, err
}

func (r *CertificateImpl) IsCancelled(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM certificate_applications 
            WHERE id = $1 AND application_status = 'Cancelled'
        )`, id).Scan(&exists)
	return exists, err
}

func (r *CertificateImpl) FindAllWithStatus(userID int64, st entity.CertificateStatus, tp string) ([]entity.CertificateApplication, error) {
	query := `
        SELECT id, student_id, application_status, certificate_type, obtain_method,
               COALESCE(comment, ''), COALESCE(rejection_reason, ''), created_at
        FROM certificate_applications
        WHERE student_id = $1 AND application_status = $2 AND certificate_type = $3
        ORDER BY created_at
		`
	rows, err := r.db.Query(query, userID, st, tp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanItems(rows)
}

func (r *CertificateImpl) FindAllByStudent(studentID int64) ([]entity.CertificateApplication, error) {
	query := `
        SELECT id, student_id, application_status, certificate_type, obtain_method,
               COALESCE(comment, ''), COALESCE(rejection_reason, ''), created_at
        FROM certificate_applications
        WHERE student_id = $1 AND application_status <> 'Cancelled'
        ORDER BY created_at DESC LIMIT 10
    `
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanItems(rows)
}

func (r *CertificateImpl) FindByID(id int64) (*entity.CertificateApplication, error) {
	row := r.db.QueryRow(
		`SELECT id, student_id, application_status, certificate_type, obtain_method,
         COALESCE(comment, ''), COALESCE(rejection_reason, ''), created_at, form_data
         FROM certificate_applications WHERE id = $1`, id)

	var found entity.CertificateApplication
	var formDataJSON []byte

	err := row.Scan(
		&found.ID,
		&found.StudentID,
		&found.Status,
		&found.Type,
		&found.ObtainMethod,
		&found.Comment,
		&found.RejectionReason,
		&found.CreatedAt,
		&formDataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if len(formDataJSON) > 0 {
		var formData map[string]interface{}
		if err := json.Unmarshal(formDataJSON, &formData); err != nil {
			found.FormData = make(map[string]interface{})
		} else {
			found.FormData = formData
		}
	} else {
		found.FormData = make(map[string]interface{})
	}

	return &found, nil
}

func (r *CertificateImpl) FindAttachmentsByOrderID(orderID int64) ([]entity.CertificateAttachment, error) {
	rows, err := r.db.Query(`
        SELECT id, order_id, file_id, uploaded_at
        FROM certificate_attachments
        WHERE order_id = $1
        ORDER BY uploaded_at DESC
    `, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []entity.CertificateAttachment
	for rows.Next() {
		var att entity.CertificateAttachment
		err := rows.Scan(
			&att.ID,
			&att.OrderID,
			&att.FileID,
			&att.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, att)
	}
	return attachments, rows.Err()
}

func scanItems(rows *sql.Rows) ([]entity.CertificateApplication, error) {
	var items []entity.CertificateApplication
	for rows.Next() {
		var item entity.CertificateApplication
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.Status,
			&item.Type,
			&item.ObtainMethod,
			&item.Comment,
			&item.RejectionReason,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
