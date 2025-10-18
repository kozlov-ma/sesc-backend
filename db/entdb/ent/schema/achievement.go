package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// Achievement holds the schema definition for the Achievement entity.
type Achievement struct {
	ent.Schema
}

// Fields of the Achievement.
func (Achievement) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Unique().
			Immutable(),
		field.UUID("owner_id", uuid.UUID{}),
		field.UUID("template_id", uuid.UUID{}),
		field.UUID("department_id", uuid.UUID{}),
		field.String("status").
			Default("draft").
			NotEmpty(),
		field.Int("points").
			Default(0),
	}
}

// Edges of the Achievement.
func (Achievement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("documents", AchievementDocument.Type),
		edge.To("reviews", AchievementReview.Type),
		edge.From("owner", User.Type).
			Ref("achievements").
			Field("owner_id").
			Unique().
			Required(),
		edge.From("departments", Department.Type).
			Ref("achievements").
			Field("department_id").
			Unique().
			Required(),
		edge.From("template", AchievementTemplate.Type).
			Ref("achievements").
			Field("template_id").
			Unique().
			Required(),
	}
}
