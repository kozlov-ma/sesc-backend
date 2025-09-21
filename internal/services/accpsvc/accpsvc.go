package accpsvc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	accountingperiod "github.com/kozlov-ma/sesc-backend/accounting_period"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAccountingPeriod "github.com/kozlov-ma/sesc-backend/db/entdb/ent/accountingperiod"
	entFile "github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
	"github.com/kozlov-ma/sesc-backend/internal/config"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
)

type FileService interface {
	Delete(ctx context.Context, id uuid.UUID) error
}

type ACCPS struct {
	client      *ent.Client
	fileService FileService
	config      *config.AccountingPeriodConfig
}

func New(client *ent.Client, fileService FileService, cfg *config.AccountingPeriodConfig) *ACCPS {
	return &ACCPS{
		client:      client,
		fileService: fileService,
		config:      cfg,
	}
}

func (s *ACCPS) CreateAccountingPeriod(
	ctx context.Context,
	opt accountingperiod.CreateOptions,
) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/create_accounting_period")
	rec.Sub("params").Set(
		"name", opt.Name,
		"description", opt.Description,
	)

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.client.AccountingPeriod.Query().
		Where(entAccountingPeriod.StatusEQ(accountingperiod.StatusPlanning)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		rec.Add(events.Error, fmt.Errorf("failed to check for existing planning period: %w", err))
		return nil, err
	}
	if existing != nil {
		rec.Add(events.Error, "planning period already exists")
		return nil, accountingperiod.ErrPlanningPeriodAlreadyExists
	}

	var startPlanningDate, startAchCollectionDate, finishDate *time.Time
	if opt.StartPlanningDate != "" {
		if t, err := time.Parse("2006-01-02", opt.StartPlanningDate); err == nil {
			startPlanningDate = &t
		}
	}
	if opt.StartAchCollectionDate != "" {
		if t, err := time.Parse("2006-01-02", opt.StartAchCollectionDate); err == nil {
			startAchCollectionDate = &t
		}
	}
	if opt.FinishDate != "" {
		if t, err := time.Parse("2006-01-02", opt.FinishDate); err == nil {
			finishDate = &t
		}
	}

	rec.Sub("create_period_record").Wrap(ctx)
	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	period, err := s.client.AccountingPeriod.Create().
		SetName(opt.Name).
		SetNillableDescription(&opt.Description).
		SetNillableStartPlanningDate(startPlanningDate).
		SetNillableStartAchievementCollectionDate(startAchCollectionDate).
		SetNillableFinishDate(finishDate).
		SetStatus(accountingperiod.StatusPlanning).
		Save(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if ent.IsValidationError(err) || ent.IsConstraintError(err) {
		return nil, accountingperiod.ErrInvalidAccountingPeriodName
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't create accounting period: %w", err))
		return nil, err
	}

	rec.Add("created_period_id", period.ID)
	return period, nil
}

func (s *ACCPS) GetAccountingPeriodByID(ctx context.Context, id int) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/accounting_period_by_id")
	rec.Add("period_id", id)

	period, err := s.client.AccountingPeriod.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "accounting period not found")
			return nil, accountingperiod.ErrAccountingPeriodNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to get accounting period: %w", err))
		return nil, err
	}

	return period, nil
}

func (s *ACCPS) GetAccountingPeriods(ctx context.Context) ([]*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/get_accounting_periods")

	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	periods, err := s.client.AccountingPeriod.Query().
		Order(ent.Desc(entAccountingPeriod.FieldStartPlanningDate)).
		All(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("failed to get accounting periods: %w", err))
		return nil, err
	}

	rec.Add("periods_count", len(periods))
	return periods, nil
}

func (s *ACCPS) GetCurrentAccountingPeriod(ctx context.Context) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/get_current_accounting_period")

	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	period, err := s.client.AccountingPeriod.Query().
		Where(entAccountingPeriod.StatusEQ(accountingperiod.StatusAchievementCollection)).
		Only(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "no active accounting period found")
			return nil, accountingperiod.ErrAccountingPeriodNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to get current accounting period: %w", err))
		return nil, err
	}

	return period, nil
}

func (s *ACCPS) GetPlanningAccountingPeriod(ctx context.Context) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/get_planning_accounting_period")

	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	period, err := s.client.AccountingPeriod.Query().
		Where(entAccountingPeriod.StatusEQ(accountingperiod.StatusPlanning)).
		Only(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if err != nil {
		if ent.IsNotFound(err) {
			rec.Add(events.Error, "no planning accounting period found")
			return nil, accountingperiod.ErrAccountingPeriodNotFound
		}
		rec.Add(events.Error, fmt.Errorf("failed to get planning accounting period: %w", err))
		return nil, err
	}

	return period, nil
}

func (s *ACCPS) UpdateAccountingPeriod(
	ctx context.Context,
	id int,
	opt accountingperiod.UpdateOptions,
) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/update_accounting_period")
	rec.Add("period_id", id)
	rec.Sub("params").Set(
		"name", opt.Name,
		"description", opt.Description,
	)

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	period, err := s.GetAccountingPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if period.Status != accountingperiod.StatusPlanning {
		rec.Add(events.Error, "period cannot be modified in current status")
		return nil, accountingperiod.ErrPeriodCannotBeModified
	}

	var startPlanningDate, startAchCollectionDate, finishDate *time.Time
	if opt.StartPlanningDate != "" {
		if t, err := time.Parse("2006-01-02", opt.StartPlanningDate); err == nil {
			startPlanningDate = &t
		}
	}
	if opt.StartAchCollectionDate != "" {
		if t, err := time.Parse("2006-01-02", opt.StartAchCollectionDate); err == nil {
			startAchCollectionDate = &t
		}
	}
	if opt.FinishDate != "" {
		if t, err := time.Parse("2006-01-02", opt.FinishDate); err == nil {
			finishDate = &t
		}
	}

	rec.Sub("update_period_record").Wrap(ctx)
	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	update := s.client.AccountingPeriod.UpdateOneID(id).
		SetName(opt.Name).
		SetNillableDescription(&opt.Description).
		SetNillableStartPlanningDate(startPlanningDate).
		SetNillableStartAchievementCollectionDate(startAchCollectionDate).
		SetNillableFinishDate(finishDate)

	updatedPeriod, err := update.Save(ctx)
	statrec.Add(events.PostgresTime, time.Since(stime))

	if ent.IsValidationError(err) || ent.IsConstraintError(err) {
		return nil, accountingperiod.ErrInvalidAccountingPeriodName
	}
	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't update accounting period: %w", err))
		return nil, err
	}

	return updatedPeriod, nil
}

func (s *ACCPS) BeginCollection(ctx context.Context, id int) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/begin_collection")
	rec.Add("period_id", id)

	period, err := s.GetAccountingPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !accountingperiod.IsValidTransition(
		accountingperiod.Status(period.Status),
		accountingperiod.StatusAchievementCollection,
	) {
		rec.Add(events.Error, "invalid status transition")
		return nil, accountingperiod.ErrInvalidStatusTransition
	}

	existing, err := s.client.AccountingPeriod.Query().
		Where(entAccountingPeriod.StatusEQ(accountingperiod.StatusAchievementCollection)).
		Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		rec.Add(events.Error, fmt.Errorf("failed to check for existing active period: %w", err))
		return nil, err
	}
	if existing != nil {
		rec.Add(events.Error, "active period already exists")
		return nil, accountingperiod.ErrActivePeriodAlreadyExists
	}

	rec.Sub("begin_collection_record").Wrap(ctx)
	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	updatedPeriod, err := s.client.AccountingPeriod.UpdateOneID(id).
		SetStatus(accountingperiod.StatusAchievementCollection).
		Save(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("couldn't begin collection: %w", err))
		return nil, err
	}

	return updatedPeriod, nil
}

func (s *ACCPS) FinishPeriod(ctx context.Context, id int) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/finish_period")
	rec.Add("period_id", id)

	period, err := s.GetAccountingPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !accountingperiod.IsValidTransition(accountingperiod.Status(period.Status), accountingperiod.StatusFinished) {
		rec.Add(
			events.Error,
			fmt.Sprintf("invalid status transition, %s -> %s", period.Status, accountingperiod.StatusFinished),
		)
		return nil, accountingperiod.ErrInvalidStatusTransition
	}

	var updatedPeriod *ent.AccountingPeriod

	err = txwrapper.WithTx(ctx, s.client, sql.LevelReadCommitted, rec, func(tx *ent.Tx) error {
		updatedPeriod, err = tx.AccountingPeriod.UpdateOneID(id).
			SetStatus(accountingperiod.StatusFinished).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("couldn't finish period: %w", err)
		}

		rec.Set("debug_period_finished", true)
		rec.Set("debug_period_id", id)

		if err := s.deleteOldDocuments(ctx, rec, tx); err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to delete old documents: %w", err))
			return err
		}
		return nil
	})

	if err != nil {
		rec.Add(events.Error, err)
		return nil, err
	}

	return updatedPeriod, nil
}

func (s *ACCPS) deleteOldDocuments(ctx context.Context, rec *event.Record, tx *ent.Tx) error {
	if s.config.PeriodsToDeleteDocuments <= 0 {
		rec.Set("delete_skipped", "periods_to_delete_documents is 0 or negative")
		return nil
	}

	finishedPeriods, err := tx.AccountingPeriod.Query().
		Where(entAccountingPeriod.StatusEQ(accountingperiod.StatusFinished)).
		Order(ent.Desc(entAccountingPeriod.FieldID)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to get finished periods: %w", err)
	}

	rec.Set("finished_periods_count", len(finishedPeriods))

	if len(finishedPeriods) <= s.config.PeriodsToDeleteDocuments {
		rec.Set("delete_skipped", "not enough finished periods to delete")
		return nil
	}

	periodsToDelete := finishedPeriods[s.config.PeriodsToDeleteDocuments:]
	rec.Set("periods_to_delete_count", len(periodsToDelete))

	var deletedFilesCount int
	var deletedFilesSize int64

	for _, period := range periodsToDelete {
		files, err := tx.File.Query().
			Where(entFile.AccountingPeriodIDEQ(period.ID)).
			All(ctx)
		if err != nil {
			rec.Add(events.Error, fmt.Errorf("failed to get files for period %d: %w", period.ID, err))
			continue
		}

		rec.Sub("delete_period_files").Set(
			"period_id", period.ID,
			"period_name", period.Name,
			"files_count", len(files),
		)

		for _, file := range files {
			if err := s.fileService.Delete(ctx, file.ID); err != nil {
				rec.Add(events.Error, fmt.Errorf("failed to delete file %s: %w", file.ID.String(), err))
				continue
			}
			deletedFilesCount++
			deletedFilesSize += int64(file.Size)
		}
	}

	rec.Set("deleted_files_count", deletedFilesCount)
	rec.Set("deleted_files_size_bytes", deletedFilesSize)

	return nil
}

func (s *ACCPS) CancelPeriod(ctx context.Context, id int) (*ent.AccountingPeriod, error) {
	return s.updatePeriodStatus(
		ctx,
		id,
		accountingperiod.StatusCancelled,
		"cancel_period",
		"couldn't cancel period",
		func(update *ent.AccountingPeriodUpdateOne) *ent.AccountingPeriodUpdateOne {
			now := time.Now().Truncate(time.Second)
			return update.SetNillableCancelDate(&now)
		},
	)
}

func (s *ACCPS) MarkAsNonExecuted(ctx context.Context, id int) (*ent.AccountingPeriod, error) {
	return s.updatePeriodStatus(
		ctx,
		id,
		accountingperiod.StatusNotExecuted,
		"mark_as_non_executed",
		"couldn't mark as non executed",
		func(update *ent.AccountingPeriodUpdateOne) *ent.AccountingPeriodUpdateOne {
			now := time.Now().Truncate(time.Second)
			return update.SetNillableBecameNonExecutedDate(&now)
		},
	)
}

func (s *ACCPS) updatePeriodStatus(
	ctx context.Context,
	id int,
	newStatus accountingperiod.Status,
	operationName string,
	errorMessage string,
	updateFunc func(*ent.AccountingPeriodUpdateOne) *ent.AccountingPeriodUpdateOne,
) (*ent.AccountingPeriod, error) {
	rec := event.Get(ctx).Sub("sesc/" + operationName)
	rec.Add("period_id", id)

	period, err := s.GetAccountingPeriodByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !accountingperiod.IsValidTransition(accountingperiod.Status(period.Status), newStatus) {
		rec.Add(
			events.Error,
			fmt.Sprintf("invalid status transition, %s -> %s", period.Status, newStatus),
		)
		return nil, accountingperiod.ErrInvalidStatusTransition
	}

	rec.Sub(operationName + "_record").Wrap(ctx)
	statrec := rec.Sub("stats")
	statrec.Add(events.PostgresQueries, 1)
	stime := time.Now()

	update := s.client.AccountingPeriod.UpdateOneID(id).SetStatus(string(newStatus))
	updatedPeriod, err := updateFunc(update).Save(ctx)

	statrec.Add(events.PostgresTime, time.Since(stime))

	if err != nil {
		rec.Add(events.Error, fmt.Errorf("%s: %w", errorMessage, err))
		return nil, err
	}

	return updatedPeriod, nil
}
