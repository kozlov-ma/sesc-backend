package schema

import (
	"context"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
	gen "github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/hook"
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
		field.Int("academic_degree").Default(0),
		field.String("academic_title").Default(""),
		field.String("honors").Default(""),
		field.String("category").Default(""),
		field.Time("date_of_employment"),
		field.Time("unemployment_date").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now),
	}
}

func (User) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return hook.UserFunc(func(ctx context.Context, m *gen.UserMutation) (ent.Value, error) {
					if len(m.Fields()) > 0 {
						m.SetUpdatedAt(time.Now())
					}
					return next.Mutate(ctx, m)
				})
			},
			ent.OpUpdate|ent.OpUpdateOne,
		),
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
