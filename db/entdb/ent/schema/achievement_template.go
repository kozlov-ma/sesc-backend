package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

// AchievementTemplate holds the schema definition for the AchievementTemplate entity.
type AchievementTemplate struct {
	ent.Schema
}

// Fields of the AchievementTemplate.
func (AchievementTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Unique().
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(600),
		field.Text("description").
			Optional(),
		field.Int("points_limit").
			Positive(),
		field.UUID("group_id", uuid.UUID{}),
		field.Bool("active").
			Default(true),
		field.Int("reviewer_role").GoType(sesc.Role(0)),
	}
}

// Edges of the AchievementTemplate.
func (AchievementTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("group", AchievementGroup.Type).
			Ref("templates").
			Field("group_id").
			Unique().
			Required(),
		edge.To("achievements", Achievement.Type),
	}
}
