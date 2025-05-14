package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
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
	}
}
