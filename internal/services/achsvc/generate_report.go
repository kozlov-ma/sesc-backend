package achsvc

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/company"
	"github.com/kozlov-ma/sesc-backend/company/companyquery"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/internal/services/txwrapper"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/xuri/excelize/v2"
)

type userReport struct {
	FullName       string
	DepartmentName string
	JobTitle       string
	TotalPoints    int
}

// ACCESSTODO
func (s *ACS) GenerateUserPointsReport(ctx context.Context) (*bytes.Buffer, error) {
	rec := event.Get(ctx).Sub("sesc/generate_user_points_report")

	statsRec := event.Get(ctx).Sub("stats")

	queryCount := 0
	startTime := time.Now()
	var excelBuffer *bytes.Buffer

	deps, err := s.company.Departments(ctx, companyquery.Departments{})
	if err != nil {
		return nil, fmt.Errorf("failed to query all departments: %w", err)
	}

	departmentNames := make(map[string]string)
	for _, d := range deps {
		departmentNames[d.ID] = d.Name
	}

	users, err := s.company.Users(ctx, companyquery.Users{})
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}

	err = txwrapper.WithTx(ctx, s.client, sql.LevelRepeatableRead, rec, func(tx *ent.Tx) error {
		reportData, err := s.calculateUserPointsData(ctx, tx, rec, users, departmentNames, &queryCount)
		if err != nil {
			return err
		}

		excelBuffer, err = s.createExcelReport(ctx, rec, reportData)
		if err != nil {
			return err
		}

		rec.Sub("result").Set(
			"users_count", len(users),
			"report_rows", len(reportData),
			"excel_size_bytes", excelBuffer.Len(),
		)
		return nil
	})

	statsRec.Add("postgres_queries", queryCount)
	statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())

	if err != nil {
		return nil, err
	}
	return excelBuffer, nil
}

func (s *ACS) calculateUserPointsData(
	ctx context.Context,
	tx *ent.Tx,
	rec *event.Record,
	users []company.User,
	departmentNames map[string]string,
	queryCount *int,
) ([]userReport, error) {
	var reportData []userReport

	err := rec.Operation("calculate_user_points", func(opRec *event.Record) error {
		for i, usr := range users {
			userRec := opRec.Sub(fmt.Sprintf("user_%d", i))
			userRec.Set("user_id", usr.ID)

			pointsSum, err := s.getUserTotalPoints(ctx, usr.ID, tx, userRec, queryCount)
			if err != nil {
				return err
			}

			reportData = append(reportData, userReport{
				FullName:       usr.FullName,
				DepartmentName: cmp.Or(departmentNames[usr.DepartmentID], "неизвестный отдел"),
				JobTitle:       usr.Extras.JobTitle,
				TotalPoints:    pointsSum,
			})

			userRec.Set("total_points", pointsSum)
		}
		return nil
	})
	return reportData, err
}

func (s *ACS) getUserTotalPoints(
	ctx context.Context,
	userID string,
	tx *ent.Tx,
	userRec *event.Record,
	queryCount *int,
) (int, error) {
	queryStart := time.Now()
	pointsSum := 0

	// First check if user has any done achievements
	count, err := tx.Achievement.Query().
		Where(
			entAchievement.OwnerID(userID),
			entAchievement.StatusEQ(string(achievement.StatusDone)),
		).
		Count(ctx)
	*queryCount++
	userRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

	if err != nil {
		userRec.Add(events.Error, fmt.Errorf("failed to count achievements for user %s: %w", userID, err))
		return 0, err
	}

	// Only sum points if there are achievements
	if count > 0 {
		sumResult, err := tx.Achievement.Query().
			Where(
				entAchievement.OwnerID(userID),
				entAchievement.StatusEQ(string(achievement.StatusDone)),
			).
			Aggregate(ent.Sum(entAchievement.FieldPoints)).
			Int(ctx)
		if err != nil {
			userRec.Add(events.Error, fmt.Errorf("failed to sum points for user %s: %w", userID, err))
			return 0, err
		}
		pointsSum = sumResult
	}

	return pointsSum, nil
}

func (s *ACS) createExcelReport(
	_ context.Context,
	rec *event.Record,
	reportData []userReport,
) (*bytes.Buffer, error) {
	var excelBuffer *bytes.Buffer
	err := rec.Operation("create_excel_file", func(opRec *event.Record) error {
		file := excelize.NewFile()
		defer file.Close()

		sheetName := "Отчет по баллам"
		index, err := file.NewSheet(sheetName)
		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to create new sheet: %w", err))
			return err
		}

		err = s.setExcelHeaders(file, sheetName, opRec)
		if err != nil {
			return err
		}

		err = s.setExcelData(file, sheetName, reportData, opRec)
		if err != nil {
			return err
		}

		file.SetActiveSheet(index)
		if err := file.DeleteSheet("Sheet1"); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to delete default sheet: %w", err))
			return err
		}

		buffer := new(bytes.Buffer)
		if err := file.Write(buffer); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to write excel file to buffer: %w", err))
			return err
		}

		excelBuffer = buffer
		opRec.Set("excel_size_bytes", buffer.Len())
		return nil
	})
	return excelBuffer, err
}

func (s *ACS) setExcelHeaders(file *excelize.File, sheetName string, opRec *event.Record) error {
	headers := []string{"ФИО", "Подразделение", "Должность", "Итоговое количество баллов", "Стоимость балла", "Сумма"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := file.SetCellValue(sheetName, cell, header); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set header cell value: %w", err))
			return err
		}
	}
	return nil
}

func (s *ACS) setExcelData(file *excelize.File, sheetName string, reportData []userReport, opRec *event.Record) error {
	for i, data := range reportData {
		row := i + 2
		if err := file.SetCellValue(sheetName, fmt.Sprintf("A%d", row), data.FullName); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set full name cell value: %w", err))
			return err
		}
		if err := file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.DepartmentName); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set subdivision cell value: %w", err))
			return err
		}
		if err := file.SetCellValue(sheetName, fmt.Sprintf("C%d", row), data.JobTitle); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set job title cell value: %w", err))
			return err
		}
		if err := file.SetCellValue(sheetName, fmt.Sprintf("D%d", row), data.TotalPoints); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set total points cell value: %w", err))
			return err
		}
		if err := file.SetCellValue(sheetName, fmt.Sprintf("E%d", row), ""); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set point cost cell value: %w", err))
			return err
		}
		if err := file.SetCellFormula(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("D%d*E%d", row, row)); err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to set formula cell value: %w", err))
			return err
		}
	}
	return nil
}
