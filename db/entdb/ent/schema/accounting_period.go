package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type AccountingPeriod struct {
	ent.Schema
}

func (AccountingPeriod) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique().
			NotEmpty(),
		field.String("description").
			Nillable().
			Optional(),
		field.Time("start_planning_date").
			Nillable().
			Optional(),
		field.Time("start_achievement_collection_date").
			Nillable().
			Optional(),
		field.Time("finish_date").
			Nillable().
			Optional(),
		field.Time("cancel_date").
			Nillable().
			Optional(),
		field.Time("became_non_executed_date").
			Nillable().
			Optional(),
		field.String("status").
			Default("planning").
			NotEmpty(),
	}
}

func (AccountingPeriod) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("files", File.Type),
	}
}

// func (AccountingPeriod) Indexes() []ent.Index {
// 	return []ent.Index{
// 		index.Fields("status"),
// 		index.Fields("name"),
// 	}
// }
