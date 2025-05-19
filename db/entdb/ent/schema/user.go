package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
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

		field.String("subdivision"),
		field.String("job_title"),
		field.Float("employment_rate").Default(1),
		field.Int("academic_degree").Optional(),
		field.Int("personnel_category"),
		field.Int("employment_type"),
		field.String("academic_title").Optional(),
		field.String("honors").Optional(),
		field.String("category").Optional(),

		field.Time("date_of_employment"),
		field.Time("unemployment_date").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
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
