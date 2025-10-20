package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// AchievementReview holds the schema definition for the AchievementReview entity.
type AchievementReview struct {
	ent.Schema
}

// Fields of the AchievementReview.
func (AchievementReview) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Unique().
			Immutable(),
		field.UUID("achievement_id", uuid.UUID{}),
		field.String("reviewer_id"),
		field.Int("points_assigned"),
		field.Text("comment").
			Optional(),
	}
}

// Edges of the AchievementReview.
func (AchievementReview) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("achievement", Achievement.Type).
			Ref("reviews").
			Field("achievement_id").
			Unique().
			Required(),
	}
}
