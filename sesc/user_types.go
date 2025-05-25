package sesc

// PersonnelCategory represents the category of personnel
type PersonnelCategory int

const (
	// PersonnelCategoryUnknown represents an unknown personnel category
	PersonnelCategoryUnknown PersonnelCategory = 0

	// PersonnelCategoryPPS represents "Профессорско-педагогический состав"
	PersonnelCategoryPPS PersonnelCategory = 1

	// PersonnelCategoryPED represents "Педагогический состав"
	PersonnelCategoryPED PersonnelCategory = 2

	// PersonnelCategoryUVP represents "Учебно-вспомогательный персонал"
	PersonnelCategoryUVP PersonnelCategory = 3

	// PersonnelCategoryAUP represents "Административно-управленческий персонал"
	PersonnelCategoryAUP PersonnelCategory = 4
)

// String returns the string representation of the personnel category
func (pc PersonnelCategory) String() string {
	switch pc {
	case PersonnelCategoryPPS:
		return "Профессорско-педагогический состав"
	case PersonnelCategoryPED:
		return "Педагогический состав"
	case PersonnelCategoryUVP:
		return "Учебно-вспомогательный персонал"
	case PersonnelCategoryAUP:
		return "Административно-управленческий персонал"
	case PersonnelCategoryUnknown:
		fallthrough
	default:
		return "Неизвестная категория"
	}
}

// AcademicDegree represents the academic degree of a user
type AcademicDegree int

const (
	// AcademicDegreeNone represents no academic degree
	AcademicDegreeNone AcademicDegree = 0

	// AcademicDegreeCandidate represents "Кандидат наук"
	AcademicDegreeCandidate AcademicDegree = 1

	// AcademicDegreeDoctor represents "Доктор наук"
	AcademicDegreeDoctor AcademicDegree = 2
)

// String returns the string representation of the academic degree
func (ad AcademicDegree) String() string {
	switch ad {
	case AcademicDegreeNone:
		return "Нет"
	case AcademicDegreeCandidate:
		return "Кандидат наук"
	case AcademicDegreeDoctor:
		return "Доктор наук"
	default:
		return "Неизвестная степень"
	}
}

// EmploymentType represents the type of employment
type EmploymentType int

const (
	// EmploymentTypeUnknown represents an unknown employment type
	EmploymentTypeUnknown EmploymentType = 0

	// EmploymentTypePrimary represents "Основное место работы"
	EmploymentTypePrimary EmploymentType = 1

	// EmploymentTypeInternalSecondary represents "Внутреннее совместительство"
	EmploymentTypeInternalSecondary EmploymentType = 2

	// EmploymentTypeExternalSecondary represents "Внешнее совместительство"
	EmploymentTypeExternalSecondary EmploymentType = 3
)

// String returns the string representation of the employment type
func (et EmploymentType) String() string {
	switch et {
	case EmploymentTypePrimary:
		return "Основное место работы"
	case EmploymentTypeInternalSecondary:
		return "Внутреннее совместительство"
	case EmploymentTypeExternalSecondary:
		return "Внешнее совместительство"
	case EmploymentTypeUnknown:
		fallthrough
	default:
		return "Неизвестный тип занятости"
	}
}
