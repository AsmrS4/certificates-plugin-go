package entity

type CertUser struct {
	ID            int64  `db:"id"`
	FullName      string `db:"full_name"`
	ExternalID    string `db:"external_id"`
	TsuAccountsID string `db:"tsu_accounts_id"`
	TsuLinked     bool   `db:"tsu_linked"`
	IsTeacher     bool   `db:"is_teacher"`
	IsStudent     bool   `db:"is_student"`
	IsDeanOffice  bool   `db:"is_dean_office"`
	UpdatedAt     string `db:"updated_at"`
}

type CertUserPosition struct {
	ID              int64  `db:"id"`
	UserID          int64  `db:"user_id"`
	PositionType    string `db:"position_type"`
	Status          string `db:"status"`
	NationalityType string `db:"nationality_type"`
	FundingType     string `db:"funding_type"`
	EducationForm   string `db:"education_form"`
	FacultyName     string `db:"faculty_name"`
	DepartmentName  string `db:"department_name"`
	ProgramName     string `db:"program_name"`
	StreamName      string `db:"stream_name"`
	GroupCode       string `db:"group_code"`
	GroupName       string `db:"group_name"`
}
