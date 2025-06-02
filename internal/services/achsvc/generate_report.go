package achsvc

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	entAchievement "github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/user"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/pkg/event/events"
	"github.com/xuri/excelize/v2"
)

type userReport struct {
	FullName    string
	Subdivision string
	JobTitle    string
	TotalPoints int
}

func (s *ACS) GenerateUserPointsReport(ctx context.Context) (*bytes.Buffer, error) {
	rec := event.Get(ctx).Sub("sesc/generate_user_points_report")

	statsRec := event.Get(ctx).Sub("stats")
	queryCount := 0
	startTime := time.Now()
	defer func() {
		statsRec.Add("postgres_queries", queryCount)
		statsRec.Add("total_time_ms", time.Since(startTime).Milliseconds())
	}()

	users, err := s.queryAllUsersForReport(ctx, rec, &queryCount)
	if err != nil {
		return nil, err
	}

	reportData, err := s.calculateUserPointsData(ctx, rec, users, &queryCount)
	if err != nil {
		return nil, err
	}

	excelBuffer, err := s.createExcelReport(ctx, rec, reportData)
	if err != nil {
		return nil, err
	}

	rec.Sub("result").Set(
		"users_count", len(users),
		"report_rows", len(reportData),
		"excel_size_bytes", excelBuffer.Len(),
	)

	return excelBuffer, nil
}

func (s *ACS) queryAllUsersForReport(ctx context.Context, rec *event.Record, queryCount *int) ([]*ent.User, error) {
	var users []*ent.User
	err := rec.Operation("query_all_users", func(opRec *event.Record) error {
		queryStart := time.Now()
		userList, err := s.client.User.Query().
			WithDepartment().
			Order(ent.Asc(user.FieldLastName), ent.Asc(user.FieldFirstName)).
			All(ctx)
		*queryCount++
		opRec.Add("query_time_ms", time.Since(queryStart).Milliseconds())

		if err != nil {
			opRec.Add(events.Error, fmt.Errorf("failed to query all users: %w", err))
			return err
		}

		users = userList
		opRec.Set("users_count", len(users))
		return nil
	})
	return users, err
}

func (s *ACS) calculateUserPointsData(
	ctx context.Context,
	rec *event.Record,
	users []*ent.User,
	queryCount *int,
) ([]userReport, error) {
	var reportData []userReport

	err := rec.Operation("calculate_user_points", func(opRec *event.Record) error {
		for i, usr := range users {
			userRec := opRec.Sub(fmt.Sprintf("user_%d", i))
			userRec.Set("user_id", usr.ID)

			pointsSum, err := s.getUserTotalPoints(ctx, usr.ID, userRec, queryCount)
			if err != nil {
				return err
			}

			fullName := s.formatUserFullName(usr)
			subdivision := s.getUserSubdivision(usr)

			reportData = append(reportData, userReport{
				FullName:    fullName,
				Subdivision: subdivision,
				JobTitle:    usr.JobTitle,
				TotalPoints: pointsSum,
			})

			userRec.Set("total_points", pointsSum)
		}
		return nil
	})
	return reportData, err
}

func (s *ACS) getUserTotalPoints(
	ctx context.Context,
	userID uuid.UUID,
	userRec *event.Record,
	queryCount *int,
) (int, error) {
	queryStart := time.Now()
	pointsSum := 0

	// First check if user has any done achievements
	count, err := s.client.Achievement.Query().
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
		sumResult, err := s.client.Achievement.Query().
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

func (s *ACS) formatUserFullName(usr *ent.User) string {
	fullName := usr.LastName
	if usr.FirstName != "" {
		fullName += " " + usr.FirstName
	}
	if usr.MiddleName != "" {
		fullName += " " + usr.MiddleName
	}
	return fullName
}

func (s *ACS) getUserSubdivision(usr *ent.User) string {
	subdivision := usr.Subdivision
	if subdivision == "" && usr.Edges.Department != nil {
		subdivision = usr.Edges.Department.Name
	}
	return subdivision
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
		if err := file.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.Subdivision); err != nil {
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
