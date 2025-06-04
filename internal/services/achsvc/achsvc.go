package achsvc

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/sesc"
)

type (
	UUID                             = uuid.UUID
	Role                             = sesc.Role
	UserUpdateOptions                = sesc.UserUpdateOptions
	AchievementGroupCreateOptions    = achievement.GroupCreateOptions
	AchievementGroupUpdateOptions    = achievement.GroupUpdateOptions
	AchievementGroupSearchOptions    = achievement.GroupSearchOptions
	AchievementTemplateCreateOptions = achievement.TemplateCreateOptions
	AchievementTemplateUpdateOptions = achievement.TemplateUpdateOptions
	AchievementTemplateSearchOptions = achievement.TemplateSearchOptions
)

type ACS struct {
	client *ent.Client
}

func New(client *ent.Client) *ACS {
	return &ACS{
		client: client,
	}
}

func withTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
