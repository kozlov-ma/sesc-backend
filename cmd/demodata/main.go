package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/api/param"
	"github.com/kozlov-ma/sesc-backend/api/respond"
	"github.com/kozlov-ma/sesc-backend/company"
)

const (
	apiURL     = "http://localhost:8080"
	adminUser  = "kozlovma"
	adminPass  = "yandexyandex"
	maxRetries = 30
	retryDelay = 1 * time.Second
)

// TemplateData holds data for creating achievement templates
type TemplateData struct {
	Name         string
	Description  string
	PointsLimit  int64
	ReviewerRole company.Role
	GroupName    string // Reference to the group name
}

// Achievement groups data
var achievementGroups = []struct {
	Name        string
	Description string
}{
	{
		Name:        "Показатель № 1 Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы",
		Description: "Сопровождение мероприятий программы развития",
	},
	{
		Name:        "Показатель № 2 Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
		Description: "Обеспечение участия и сопровождение обучающихся в мероприятиях",
	},
	{
		Name:        "Показатель № 3 Сопровождение дистанционного курса",
		Description: "Сопровождение дистанционного курса"},
	{
		Name:        "Показатель № 4.1 Научные, научно-методические публикации работников",
		Description: "Научные и научно-методические публикации",
	},
	{
		Name:        "Показатель № 4.2 Участие в конференции с докладом без последующей публикации в сборниках трудов",
		Description: "Участие в конференции с докладом",
	},
	{
		Name:        "Показатель № 4.3 Методические пособия, рабочие программы, курсы для обучающихся СУНЦ",
		Description: "Методические пособия и программы",
	},
	{
		Name:        "Показатель № 5 Региональные предметные комиссии по проверке развёрнутых ответов участников государственной итоговой аттестации",
		Description: "Участие в региональных предметных комиссиях",
	},
}

var achievementTemplates = []TemplateData{
	// Group 1 templates
	{
		Name:         "Организация мероприятия федерального уровня",
		Description:  "Организация и проведение мероприятия федерального уровня",
		PointsLimit:  100,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 1 Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы",
	},
	{
		Name:         "Организация регионального мероприятия",
		Description:  "Организация и проведение мероприятия регионального уровня",
		PointsLimit:  50,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 1 Сопровождение (подготовка/организация, проведение) мероприятий программы развития, плана работы",
	},

	// Group 2 templates
	{
		Name:         "Сопровождение обучающихся на мероприятие федерального уровня",
		Description:  "Сопровождение обучающихся на федеральное мероприятие",
		PointsLimit:  30,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 2 Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
	},
	{
		Name:         "Сопровождение обучающихся на региональное мероприятие",
		Description:  "Сопровождение обучающихся на региональное мероприятие",
		PointsLimit:  15,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 2 Обеспечение участия в мероприятиях и сопровождение обучающихся СУНЦ УрФУ",
	},

	// Group 3 templates
	{
		Name:         "Ведение дистанционного курса",
		Description:  "Сопровождение и ведение дистанционного курса для обучающихся",
		PointsLimit:  20,
		ReviewerRole: company.ScientificDeputy,
		GroupName:    "Показатель № 3 Сопровождение дистанционного курса",
	},

	// Group 4.1 templates
	{
		Name:         "Публикация в журнале ВАК",
		Description:  "Публикация статьи в журнале из списка ВАК",
		PointsLimit:  100,
		ReviewerRole: company.ScientificDeputy,
		GroupName:    "Показатель № 4.1 Научные, научно-методические публикации работников",
	},
	{
		Name:         "Публикация в РИНЦ",
		Description:  "Публикация статьи в журнале, индексируемом в РИНЦ",
		PointsLimit:  50,
		ReviewerRole: company.OlympiadDeputy,
		GroupName:    "Показатель № 4.1 Научные, научно-методические публикации работников",
	},

	// Group 4.2 templates
	{
		Name:         "Доклад на международной конференции",
		Description:  "Выступление с докладом на международной конференции",
		PointsLimit:  40,
		ReviewerRole: company.ScientificDeputy,
		GroupName:    "Показатель № 4.2 Участие в конференции с докладом без последующей публикации в сборниках трудов",
	},
	{
		Name:         "Доклад на всероссийской конференции",
		Description:  "Выступление с докладом на всероссийской конференции",
		PointsLimit:  30,
		ReviewerRole: company.ScientificDeputy,
		GroupName:    "Показатель № 4.2 Участие в конференции с докладом без последующей публикации в сборниках трудов",
	},

	// Group 4.3 templates
	{
		Name:         "Методическое пособие",
		Description:  "Создание методического пособия для обучающихся",
		PointsLimit:  30,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 4.3 Методические пособия, рабочие программы, курсы для обучающихся СУНЦ",
	},
	{
		Name:         "Рабочая программа курса",
		Description:  "Разработка рабочей программы учебного курса",
		PointsLimit:  20,
		ReviewerRole: company.DevelopmentDeputy,
		GroupName:    "Показатель № 4.3 Методические пособия, рабочие программы, курсы для обучающихся СУНЦ",
	},
}

func main() {
	log.Println("Starting demodata population...")

	// Wait for API to be ready
	if !waitForAPI() {
		log.Fatal("API is not available")
	}

	// Login as admin
	token, err := login()
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	log.Println("Successfully logged in as admin")

	// Create achievement groups
	groupIDMap := make(map[string]string) // maps group name to ID
	for _, group := range achievementGroups {
		req := param.CreateAchievementGroupRequest{
			Name:        group.Name,
			Description: group.Description,
		}
		groupResp, err := createAchievementGroup(token, &req)
		if err != nil {
			log.Printf("Failed to create achievement group '%s': %v", group.Name, err)
			continue
		}
		groupIDMap[group.Name] = groupResp.ID.String()
		log.Printf("Created achievement group: %s (ID: %s)", group.Name, groupResp.ID.String())
	}

	// Create achievement templates
	for _, template := range achievementTemplates {
		groupID, ok := groupIDMap[template.GroupName]
		if !ok {
			log.Printf("Group '%s' not found for template '%s'", template.GroupName, template.Name)
			continue
		}

		// parse group ID string to uuid
		gid, err := uuid.FromString(groupID)
		if err != nil {
			log.Printf("Invalid group ID '%s' for template '%s': %v", groupID, template.Name, err)
			continue
		}

		req := param.CreateAchievementTemplateRequest{
			Name:         template.Name,
			Description:  template.Description,
			PointsLimit:  int(template.PointsLimit),
			GroupID:      gid,
			ReviewerRole: string(template.ReviewerRole),
		}

		err = createAchievementTemplate(token, &req)
		if err != nil {
			log.Printf("Failed to create template '%s': %v", template.Name, err)
			continue
		}
		log.Printf("Created achievement template: %s", template.Name)
	}

	log.Println("Demodata population completed successfully!")
}

type APITokenResponse struct {
	Token *string `json:"token,omitempty"`
}

func waitForAPI() bool {
	ctx := context.Background()
	for i := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/up", nil)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			time.Sleep(retryDelay)
			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return true
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		log.Printf("Waiting for API to be ready... (attempt %d/%d)", i+1, maxRetries)
		time.Sleep(retryDelay)
	}
	return false
}

func login() (string, error) {
	reqBody := map[string]string{
		"username": adminUser,
		"password": adminPass,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login request: %w", err)
	}

	ctx := context.Background()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/auth/login", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp APITokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	if loginResp.Token == nil {
		return "", errors.New("token is nil in response")
	}

	return *loginResp.Token, nil
}

func createAchievementGroup(
	token string,
	req *param.CreateAchievementGroupRequest,
) (*respond.AchievementGroup, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx := context.Background()
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL+"/achievement-groups",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var groupResp respond.AchievementGroup
	if err := json.NewDecoder(resp.Body).Decode(&groupResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &groupResp, nil
}

func createAchievementTemplate(token string, req *param.CreateAchievementTemplateRequest) error {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx := context.Background()
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		apiURL+"/achievement-templates",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		// Check if it's a duplicate error (which is okay)
		if resp.StatusCode == http.StatusConflict || strings.Contains(string(body), "already exists") {
			return nil // Ignore duplicates
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
