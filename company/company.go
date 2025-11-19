package company

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/mitchellh/hashstructure"
)

type Role string

const (
	Teacher           Role = "teacher"
	Dephead           Role = "dephead"
	ScientificDeputy  Role = "scientific_deputy"
	DevelopmentDeputy Role = "development_deputy"
	OlympiadDeputy    Role = "olympiad_deputy"
	AcademicDirector  Role = "academic_director"
	ChiefEconomist    Role = "chief_economist"
	Admin             Role = "admin"
)

var Roles = []Role{
	Teacher,
	Dephead,
	ScientificDeputy,
	DevelopmentDeputy,
	OlympiadDeputy,
	AcademicDirector,
	ChiefEconomist,
	Admin,
}

func AsRole(s string) (Role, error) {
	switch Role(s) {
	case Teacher:
		return Teacher, nil
	case Dephead:
		return Dephead, nil
	case ScientificDeputy:
		return ScientificDeputy, nil
	case DevelopmentDeputy:
		return DevelopmentDeputy, nil
	case OlympiadDeputy:
		return OlympiadDeputy, nil
	case AcademicDirector:
		return AcademicDirector, nil
	case ChiefEconomist:
		return ChiefEconomist, nil
	case Admin:
		return Admin, nil
	default:
		return "", ErrInvalidRole
	}
}

func (r Role) Name() string {
	switch r {
	case Teacher:
		return "Преподаватель"
	case Dephead:
		return "Заведующий подразделением"
	case ScientificDeputy:
		return "Заместитель директора по научной работе"
	case OlympiadDeputy:
		return "Заместитель директора по олимпиадной работе"
	case DevelopmentDeputy:
		return "Заместитель директора по развитию"
	case AcademicDirector:
		return "Академический директор"
	case ChiefEconomist:
		return "Ведущий экономист"
	case Admin:
		return "Администратор"
	case "":
		return "Нет роли"
	}

	return string(r)
}

func (r Role) String() string {
	return string(r)
}

type Department struct {
	ID          string
	Name        string
	Description string
}

type UserExtras struct {
	AcademicDegree    string
	AcademicTitle     string
	Category          string
	DateOfEmployment  string
	EmploymentRate    string
	EmploymentType    string
	Honors            string
	JobTitle          string
	PersonnelCategory string
}

type User struct {
	ID           string
	FullName     string
	PictureURL   string
	DepartmentID string
	Roles        []Role
	Extras       UserExtras
}

type Action interface {
	AllowsUser(u User) bool
}

func (u User) HasRole(rr ...Role) bool {
	for _, r := range rr {
		if slices.Contains(u.Roles, r) {
			return true
		}
	}
	return false
}

func (u User) RolesIn(rr ...Role) []Role {
	var res []Role
	for _, r := range rr {
		if slices.Contains(u.Roles, r) {
			res = append(res, r)
		}
	}
	return res
}

func (u User) HasAllRoles(rr ...Role) bool {
	for _, r := range rr {
		if !slices.Contains(u.Roles, r) {
			return false
		}
	}
	return true
}

func (u User) RoleStrings() []string {
	var ss []string
	for _, r := range u.Roles {
		ss = append(ss, string(r))
	}
	slices.Sort(ss)
	return ss
}

func (u User) Can(do Action) bool {
	return do.AllowsUser(u)
}

func (u User) Hash() string {
	h, err := hashstructure.Hash(u, &hashstructure.HashOptions{
		SlicesAsSets: true,
	})
	if err != nil {
		// This is a 'safe' panic since it will be caught in tests if it will ever appear.
		panic(fmt.Sprintf("failed to hash a company.User: %s", err.Error()))
	}

	return strconv.FormatUint(h, 16)
}

func ExEmployee(id string) User {
	return User{
		ID:       id,
		FullName: "Бывший Сотрудник",
	}
}
