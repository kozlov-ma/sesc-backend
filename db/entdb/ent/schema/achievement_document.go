package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
)

// AchievementDocument holds the schema definition for the AchievementDocument entity.
type AchievementDocument struct {
	ent.Schema
}

// Fields of the AchievementDocument.
func (AchievementDocument) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(func() uuid.UUID { return uuid.Must(uuid.NewV7()) }).
			Unique().
			Immutable(),
		field.UUID("achievement_id", uuid.UUID{}),
		field.String("name").
			NotEmpty().
			MaxLen(255),
		field.UUID("file_id", uuid.UUID{}),
		field.String("status").
			Default(achievement.DocumentStatusActive).
			NotEmpty(),
		field.Time("scheduled_deletion_at").Optional().Nillable(),
	}
}

// Edges of the AchievementDocument.
func (AchievementDocument) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("achievement", Achievement.Type).
			Ref("documents").
			Field("achievement_id").
			Unique().
			Required(),
		edge.From("file", File.Type).
			Ref("achievement_documents").
			Field("file_id").
			Unique().
			Required(),
	}
}
