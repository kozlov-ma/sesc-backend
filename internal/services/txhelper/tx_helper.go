package txhelper

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

func WithTx(
	ctx context.Context,
	client *ent.Client,
	isolation sql.IsolationLevel,
	rec *event.Record,
	fn func(tx *ent.Tx) error) error {
	txrec := rec.Sub("pg_transaction")
	txrec.Set("rollback", false)
	txrec.Set("isolation", isolation.String())
	txrec.Set("tx_begin", time.Now())

	tx, err := client.BeginTx(ctx, &sql.TxOptions{
		Isolation: isolation,
	})
	if err != nil {
		txrec.Add(events.Error, err)
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			txrec.Set("rollback", true)
			txrec.Set("tx_rollback", time.Now())
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
			txrec.Set("rollback", true)
			txrec.Set("tx_rollback", time.Now())
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		err := fmt.Errorf("failed to commit transaction: %w", err)
		txrec.Add(events.Error, err)
		return err
	}
	txrec.Set("tx_commit", time.Now())

	return nil
}
