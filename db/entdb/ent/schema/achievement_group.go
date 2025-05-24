package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// AchievementGroup holds the schema definition for the AchievementGroup entity.
type AchievementGroup struct {
	ent.Schema
}

// Fields of the AchievementGroup.
func (AchievementGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.Text("description").
			Optional(),
		field.Bool("active").
			Default(true),
	}
}

// Edges of the AchievementGroup.
func (AchievementGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("templates", AchievementTemplate.Type),
	}
}
