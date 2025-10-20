package companyquery

// User is a query a successful execution of which returns a company.User
// or a company.ErrUserNotFound.
//
// Field 'ID' is matched exactly.
// Field 'Password' is matched exactly if it is set.
type User struct {
	ID       string
	Password string
}

// Users is a query a successful execution of which returns a slice of company.User.
//
// Fields ending with 'ID' are matched exactly if present.
// Other fields are matched as a case-insensitive substring;
// if any of them matches, the user is added to the output.
type Users struct {
	DepartmentID string
	RoleID       string

	Department string
	FullName   string
	RoleName   string
}

// Department is a query a successful execution of which returns a company.Department
// or a company.ErrDepartmentNotFound.
//
// Field 'ID' is matched exactly and must be present.
type Department struct {
	ID string
}

// Departments is a query a successful execution of which returns a slice
// of company.Department.
//
// Fields ending with 'ID' are matched exactly if present.
// Other fields are matched as a case-insensitive substring.
type Departments struct {
	Name string
}
