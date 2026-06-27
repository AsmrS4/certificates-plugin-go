package repo

import (
	"database/sql"
	"fmt"
	"time"

	repository "github.com/AsmrS4/certificates-plugin-go/internal/persistence"
	"github.com/AsmrS4/certificates-plugin-go/internal/persistence/entity"
)

type UserRepoImpl struct {
	db *sql.DB
}

var _ repository.UserRepo = (*UserRepoImpl)(nil)

func NewUserRepo(db *sql.DB) *UserRepoImpl {
	return &UserRepoImpl{db: db}
}

func (r *UserRepoImpl) SaveOrUpdateUser(user *entity.CertUser, positions []entity.CertUserPosition) (bool, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var existing entity.CertUser
	var exists bool
	err = tx.QueryRow(`SELECT id, full_name, external_id, tsu_accounts_id, tsu_linked, 
        is_teacher, is_student, is_dean_office, updated_at FROM cert_users WHERE id = $1`, user.ID).Scan(
		&existing.ID, &existing.FullName, &existing.ExternalID, &existing.TsuAccountsID,
		&existing.TsuLinked, &existing.IsTeacher, &existing.IsStudent, &existing.IsDeanOffice,
		&existing.UpdatedAt,
	)
	if err == nil {
		exists = true
	} else if err != sql.ErrNoRows {
		return false, fmt.Errorf("check user exists: %w", err)
	}

	changed := false
	if exists {
		if existing.FullName != user.FullName ||
			existing.ExternalID != user.ExternalID ||
			existing.TsuAccountsID != user.TsuAccountsID ||
			existing.TsuLinked != user.TsuLinked ||
			existing.IsTeacher != user.IsTeacher ||
			existing.IsStudent != user.IsStudent ||
			existing.IsDeanOffice != user.IsDeanOffice {
			changed = true
		}
	} else {
		changed = true
	}

	if changed {
		if exists {
			_, err = tx.Exec(`UPDATE cert_users SET full_name = $1, external_id = $2, tsu_accounts_id = $3,
                tsu_linked = $4, is_teacher = $5, is_student = $6, is_dean_office = $7, updated_at = $8
                WHERE id = $9`,
				user.FullName, user.ExternalID, user.TsuAccountsID, user.TsuLinked,
				user.IsTeacher, user.IsStudent, user.IsDeanOffice, time.Now(), user.ID)
		} else {
			_, err = tx.Exec(`INSERT INTO cert_users(id, full_name, external_id, tsu_accounts_id,
                tsu_linked, is_teacher, is_student, is_dean_office, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				user.ID, user.FullName, user.ExternalID, user.TsuAccountsID, user.TsuLinked,
				user.IsTeacher, user.IsStudent, user.IsDeanOffice, time.Now())
		}
		if err != nil {
			return false, fmt.Errorf("save user: %w", err)
		}

		_, err = tx.Exec(`DELETE FROM cert_user_positions WHERE user_id = $1`, user.ID)
		if err != nil {
			return false, fmt.Errorf("delete old positions: %w", err)
		}

		for _, pos := range positions {
			_, err = tx.Exec(`INSERT INTO cert_user_positions(user_id, position_type, status,
                nationality_type, funding_type, education_form, faculty_name, department_name,
                program_name, stream_name, group_code, group_name)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
				user.ID, pos.PositionType, pos.Status, pos.NationalityType, pos.FundingType,
				pos.EducationForm, pos.FacultyName, pos.DepartmentName, pos.ProgramName,
				pos.StreamName, pos.GroupCode, pos.GroupName)
			if err != nil {
				return false, fmt.Errorf("insert position: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}

	return changed, nil
}
