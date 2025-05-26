package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).Unique(),
		field.String("first_name"),
		field.String("last_name"),
		field.String("middle_name").Default(""),
		field.String("picture_url").Optional(),
		field.Bool("suspended").Default(false),
		field.UUID("department_id", uuid.UUID{}).Optional().Nillable(),
		field.Int32("role_id"),

		field.String("subdivision").Default(""),
		field.String("job_title").Default(""),
		field.Float("employment_rate").Default(1),
		field.Int("academic_degree").GoType(sesc.AcademicDegree(0)).Optional(),
		field.Int("personnel_category").GoType(sesc.PersonnelCategory(0)).Default(0),
		field.Int("employment_type").GoType(sesc.EmploymentType(0)).Default(0),
		field.String("academic_title").Optional(),
		field.String("honors").Optional(),
		field.String("category").Optional(),

		field.Time("date_of_employment").Default(time.Now),
		field.Time("unemployment_date").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("department", Department.Type).
			Ref("users").
			Field("department_id").
			Unique(),

		edge.To("auth", AuthUser.Type).
			Unique().
			Annotations(entsql.OnDelete(entsql.Cascade)),

		edge.To("files", File.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.To("achievements", Achievement.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("reviews", AchievementReview.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("department_id"),
		index.Fields("role_id"),
		index.Fields("first_name", "last_name"),
	}
}
