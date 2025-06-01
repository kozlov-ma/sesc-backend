package sesc

type Role int

const (
	Teacher Role = iota + 1
	Dephead
	ScientificDeputy
	DevelopmentDeputy
	OlympiadDeputy
	AcademicDirector
	ChiefEconomist
)

var Roles = []Role{
	Teacher,
	Dephead,
	ScientificDeputy,
	DevelopmentDeputy,
	OlympiadDeputy,
	AcademicDirector,
	ChiefEconomist,
}

func (r Role) Name() string {
	switch r {
	case Teacher:
		return "Преподаватель"
	case Dephead:
		return "Заведующий кафедрой"
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
	default:
		return "Нет роли"
	}
}

func (r Role) String() string {
	switch r {
	case Teacher:
		return "teacher"
	case Dephead:
		return "dephead"
	case ScientificDeputy:
		return "scientific_deputy"
	case DevelopmentDeputy:
		return "development_deputy"
	case OlympiadDeputy:
		return "olympiad_deputy"
	case AcademicDirector:
		return "academic_director"
	case ChiefEconomist:
		return "chief_economist"
	default:
		return "unknown_role"
	}
}

func FromString(s string) (Role, bool) {
	switch s {
	case Teacher.String():
		return Teacher, true
	case Dephead.String():
		return Dephead, true
	case ScientificDeputy.String():
		return ScientificDeputy, true
	case DevelopmentDeputy.String():
		return DevelopmentDeputy, true
	case OlympiadDeputy.String():
		return OlympiadDeputy, true
	case AcademicDirector.String():
		return AcademicDirector, true
	case ChiefEconomist.String():
		return ChiefEconomist, true
	default:
		return 0, false
	}
}

func ValidateRole[R ~int](r R) error {
	switch Role(r) {
	case Teacher, Dephead, ScientificDeputy, DevelopmentDeputy, OlympiadDeputy, AcademicDirector, ChiefEconomist:
		return nil
	default:
		return ErrInvalidRole
	}
}
