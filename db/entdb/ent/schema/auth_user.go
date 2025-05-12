package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// AuthUser holds the schema definition for the AuthUser entity.
type AuthUser struct {
	ent.Schema
}

// Fields of the AuthUser.
func (AuthUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			Unique().
			NotEmpty(),
		field.String("password").
			NotEmpty(),
		field.UUID("auth_id", uuid.UUID{}).
			Unique(),
		field.UUID("user_id", uuid.UUID{}).
			Unique(),
	}
}

// Edges of the AuthUser.
func (AuthUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("auth").
			Field("user_id").
			Unique().
			Required(),
	}
}
