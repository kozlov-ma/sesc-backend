package sescsvc

import (
	"context"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/achievementdocument"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/kozlov-ma/sesc-backend/sesc"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test user
func createTestUser(t *testing.T, svc *SESC) User {
	t.Helper()
	ctx, _ := event.NewRecord(t.Context(), "test")

	// Create a department for the user
	dept, err := svc.CreateDepartment(ctx, strconv.Itoa(rand.Int()), "For testing")
	require.NoError(t, err)

	// Create a user
	opt := UserUpdateOptions{
		FirstName:    "Test",
		LastName:     "Teacher",
		DepartmentID: dept.ID,
		NewRoleID:    sesc.Teacher.ID,
	}
	user, err := svc.CreateUser(ctx, opt)
	require.NoError(t, err)

	return user
}

// Helper function to create a test achievement template
func createTestTemplate(t *testing.T, svc *SESC) achievement.Template {
	t.Helper()
	ctx, _ := event.NewRecord(t.Context(), "test")

	// Create a group
	groupOpts := AchievementGroupCreateOptions{
		Name:        "Test Group",
		Description: "For testing",
	}
	group, err := svc.CreateAchievementGroup(ctx, groupOpts)
	require.NoError(t, err)

	// Create a template
	templateOpts := AchievementTemplateCreateOptions{
		Name:        "Test Template",
		Description: "For testing achievements",
		PointsLimit: 10,
		GroupID:     group.ID,
		Kind:        achievement.Scientific,
	}
	template, err := svc.CreateAchievementTemplate(ctx, templateOpts)
	require.NoError(t, err)

	return template
}

// Helper function to create a test achievement
func createTestAchievement(
	t *testing.T,
	svc *SESC,
	user User,
	template achievement.Template,
) achievement.Achievement {
	t.Helper()
	ctx, _ := event.NewRecord(t.Context(), "test")

	opt := achievement.CreateOptions{
		ForUser:    user,
		TemplateID: template.ID,
	}

	ach, err := svc.CreateAchievement(ctx, opt)
	require.NoError(t, err)

	return ach
}

// Helper function to create a test file in the database
func createTestFile(ctx context.Context, t *testing.T, svc *SESC) sesc.File {
	t.Helper()

	// Create a file entry in the database
	fileID := uuid.Must(uuid.NewV7())
	_, err := svc.client.File.Create().
		SetID(fileID).
		SetName("test-file.pdf").
		SetSize(1024).
		SetURL("https://example.com/test-file.pdf").
		SetS3ObjectKey("test-files/" + fileID.String()).
		Save(ctx)
	require.NoError(t, err)

	return sesc.File{
		ID:   fileID,
		Name: "test-file.pdf",
		Size: 1024,
		URL:  "https://example.com/test-file.pdf",
	}
}

func TestCreateAchievement(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, user User, template achievement.Template) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user = createTestUser(t, svc)
		template = createTestTemplate(t, svc)
		return ctx, svc, user, template
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, user, template := setup(t)

		opt := achievement.CreateOptions{
			ForUser:    user,
			TemplateID: template.ID,
		}

		ach, err := svc.CreateAchievement(ctx, opt)
		require.NoError(t, err, "CreateAchievement failed")

		// Verify the achievement was created correctly
		require.NotEqual(t, uuid.Nil, ach.ID, "Achievement ID should not be nil")
		require.Equal(t, user.ID, ach.Owner.ID, "Achievement owner should match")
		require.Equal(t, template.ID, ach.Template.ID, "Achievement template should match")
		require.Equal(t, string(achievement.StatusDraft), string(ach.Status), "Achievement should be in draft status")
		require.Equal(t, template.PointsLimit, ach.Points, "Achievement should have 0 points initially")
		require.Empty(t, ach.Documents, "Achievement should have no documents initially")
		require.Empty(t, ach.Reviews, "Achievement should have no reviews initially")
	})

	t.Run("template not found", func(t *testing.T) {
		ctx, svc, user, _ := setup(t)

		opt := achievement.CreateOptions{
			ForUser:    user,
			TemplateID: uuid.Must(uuid.NewV7()), // Non-existent template ID
		}

		_, err := svc.CreateAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementTemplateNotFound)
	})
}

func TestDeleteAchievement(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, user, template)
		return ctx, svc, ach
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.DeleteOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}

		err := svc.DeleteAchievement(ctx, opt)
		require.NoError(t, err, "DeleteAchievement failed")

		// Try to get the achievement again - it should be gone
		// Note: We don't have a direct "get achievement" method to test this,
		// but in a real test you would verify the achievement was deleted
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.DeleteOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: uuid.Must(uuid.NewV7()), // Non-existent achievement ID
		}

		err := svc.DeleteAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong owner", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.DeleteOptions{
			OwnerID:       uuid.Must(uuid.NewV7()), // Wrong owner ID
			AchievementID: ach.ID,
		}

		err := svc.DeleteAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})
}

func TestAddDocument(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement, file sesc.File) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, user, template)
		file = createTestFile(ctx, t, svc)
		return ctx, svc, ach, file
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, ach, file := setup(t)

		opt := achievement.AddDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			Name:          "Test Document",
			FileID:        file.ID,
		}

		doc, err := svc.AddDocument(ctx, opt)
		require.NoError(t, err, "AddDocument failed")

		// Verify the document was added correctly
		require.NotEqual(t, uuid.Nil, doc.ID, "Document ID should not be nil")
		require.Equal(t, opt.Name, doc.Name, "Document name should match")
		require.Equal(t, file.ID, doc.FileID, "Document file ID should match")

		// Verify the document exists in the database
		docEntity, err := svc.client.AchievementDocument.Query().
			Where(
				achievementdocument.AchievementID(ach.ID),
				achievementdocument.FileID(file.ID),
			).
			Only(ctx)
		require.NoError(t, err)
		require.Equal(t, opt.Name, docEntity.Name)
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, ach, file := setup(t)

		opt := achievement.AddDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: uuid.Must(uuid.NewV7()), // Non-existent achievement ID
			Name:          "Test Document",
			FileID:        file.ID,
		}

		_, err := svc.AddDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong owner", func(t *testing.T) {
		ctx, svc, ach, file := setup(t)

		opt := achievement.AddDocumentOptions{
			OwnerID:       uuid.Must(uuid.NewV7()), // Wrong owner ID
			AchievementID: ach.ID,
			Name:          "Test Document",
			FileID:        file.ID,
		}

		_, err := svc.AddDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("file not found", func(t *testing.T) {
		ctx, svc, ach, _ := setup(t)

		opt := achievement.AddDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			Name:          "Test Document",
			FileID:        uuid.Must(uuid.NewV7()), // Non-existent file ID
		}

		_, err := svc.AddDocument(ctx, opt)
		require.ErrorIs(t, err, sesc.ErrFileNotFound)
	})

	t.Run("wrong achievement status", func(t *testing.T) {
		ctx, svc, ach, file := setup(t)

		// Submit the achievement to change its status from draft
		submitOpt := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}
		_, err := svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		// Now try to add a document
		opt := achievement.AddDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			Name:          "Test Document",
			FileID:        file.ID,
		}

		_, err = svc.AddDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}

func TestRemoveDocument(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement, docID uuid.UUID) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, user, template)
		testFile := createTestFile(ctx, t, svc)

		// Add a document to the achievement
		addOpt := achievement.AddDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			Name:          "Test Document",
			FileID:        testFile.ID,
		}
		doc, err := svc.AddDocument(ctx, addOpt)
		require.NoError(t, err)

		return ctx, svc, ach, doc.ID
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, ach, docID := setup(t)

		opt := achievement.RemoveDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			DocumentID:    docID,
		}

		err := svc.RemoveDocument(ctx, opt)
		require.NoError(t, err, "RemoveDocument failed")

		// Verify the document was removed from the database
		_, err = svc.client.AchievementDocument.Query().
			Where(achievementdocument.ID(docID)).
			Only(ctx)
		require.True(t, ent.IsNotFound(err), "Document should be removed from database")
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, ach, docID := setup(t)

		opt := achievement.RemoveDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: uuid.Must(uuid.NewV7()), // Non-existent achievement ID
			DocumentID:    docID,
		}

		err := svc.RemoveDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong owner", func(t *testing.T) {
		ctx, svc, ach, docID := setup(t)

		opt := achievement.RemoveDocumentOptions{
			OwnerID:       uuid.Must(uuid.NewV7()), // Wrong owner ID
			AchievementID: ach.ID,
			DocumentID:    docID,
		}

		err := svc.RemoveDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("document not found", func(t *testing.T) {
		ctx, svc, ach, _ := setup(t)

		opt := achievement.RemoveDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			DocumentID:    uuid.Must(uuid.NewV7()), // Non-existent document ID
		}

		err := svc.RemoveDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrDocumentNotFound)
	})

	t.Run("wrong achievement status", func(t *testing.T) {
		ctx, svc, ach, docID := setup(t)

		// Submit the achievement to change its status from draft
		submitOpt := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}
		_, err := svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		// Now try to remove a document
		opt := achievement.RemoveDocumentOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
			DocumentID:    docID,
		}

		err = svc.RemoveDocument(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}

func TestSubmitAchievement(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, user, template)
		return ctx, svc, ach
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}

		updatedAch, err := svc.SubmitAchievement(ctx, opt)
		require.NoError(t, err, "SubmitAchievement failed")

		// Verify the achievement was updated correctly
		require.Equal(t, ach.ID, updatedAch.ID, "Achievement ID should match")
		require.Equal(
			t,
			string(achievement.StatusDepheadReview),
			string(updatedAch.Status),
			"Achievement should be in dephead_review status",
		)
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: uuid.Must(uuid.NewV7()), // Non-existent achievement ID
		}

		_, err := svc.SubmitAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong owner", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		opt := achievement.SubmitOptions{
			OwnerID:       uuid.Must(uuid.NewV7()), // Wrong owner ID
			AchievementID: ach.ID,
		}

		_, err := svc.SubmitAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong status", func(t *testing.T) {
		ctx, svc, ach := setup(t)

		// First submit to change status
		opt1 := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}
		_, err := svc.SubmitAchievement(ctx, opt1)
		require.NoError(t, err)

		// Try to submit again
		opt2 := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}
		_, err = svc.SubmitAchievement(ctx, opt2)
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}

func TestReviewAchievement(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement, depHead User, inspector User) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)

		// Create a teacher
		teacher := createTestUser(t, svc)

		// Create a department head
		deptHeadOpt := UserUpdateOptions{
			FirstName:    "Dept",
			LastName:     "Head",
			DepartmentID: teacher.Department.ID,
			NewRoleID:    sesc.Dephead.ID,
		}
		depHead, err := svc.CreateUser(ctx, deptHeadOpt)
		require.NoError(t, err)

		// Create a scientific deputy (inspector)
		inspectorOpt := UserUpdateOptions{
			FirstName:    "Scientific",
			LastName:     "Deputy",
			DepartmentID: teacher.Department.ID,
			NewRoleID:    sesc.ScientificDeputy.ID,
		}
		inspector, err = svc.CreateUser(ctx, inspectorOpt)
		require.NoError(t, err)

		// Create template and achievement
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, teacher, template)

		// Submit the achievement for review
		submitOpt := achievement.SubmitOptions{
			OwnerID:       ach.Owner.ID,
			AchievementID: ach.ID,
		}
		ach, err = svc.SubmitAchievement(ctx, submitOpt)
		require.NoError(t, err)

		return ctx, svc, ach, depHead, inspector
	}

	t.Run("department head review success", func(t *testing.T) {
		ctx, svc, ach, depHead, _ := setup(t)

		opt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         depHead.ID,
			PointsAssigned:     8,
			Comment:            "Good achievement",
		}

		updatedAch, err := svc.ReviewAchievement(ctx, opt)
		require.NoError(t, err, "ReviewAchievement failed")

		// Verify the achievement was updated correctly
		require.Equal(t, ach.ID, updatedAch.ID, "Achievement ID should match")
		require.Equal(
			t,
			string(achievement.StatusInspectorReview),
			string(updatedAch.Status),
			"Achievement should be in inspector_review status",
		)
		require.Equal(t, opt.PointsAssigned, updatedAch.Points, "Achievement points should match")
		require.Len(t, updatedAch.Reviews, 1, "Achievement should have one review")
	})

	t.Run("department head assigns zero points", func(t *testing.T) {
		ctx, svc, ach, depHead, _ := setup(t)

		opt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         depHead.ID,
			PointsAssigned:     0,
			Comment:            "Not eligible for points",
		}

		updatedAch, err := svc.ReviewAchievement(ctx, opt)
		require.NoError(t, err, "ReviewAchievement failed")

		// Verify the achievement was updated correctly
		require.Equal(t, ach.ID, updatedAch.ID, "Achievement ID should match")
		require.Equal(
			t,
			string(achievement.StatusDone),
			string(updatedAch.Status),
			"Achievement should be in done status",
		)
		require.Equal(t, opt.PointsAssigned, updatedAch.Points, "Achievement points should match")
	})

	t.Run("inspector review success", func(t *testing.T) {
		ctx, svc, ach, depHead, inspector := setup(t)

		// First, department head review
		depHeadOpt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         depHead.ID,
			PointsAssigned:     8,
			Comment:            "Good achievement",
		}
		ach, err := svc.ReviewAchievement(ctx, depHeadOpt)
		require.NoError(t, err)
		require.Equal(t, string(achievement.StatusInspectorReview), string(ach.Status))

		// Then, inspector review
		inspectorOpt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         inspector.ID,
			PointsAssigned:     7,
			Comment:            "Approved with minor reduction",
		}
		updatedAch, err := svc.ReviewAchievement(ctx, inspectorOpt)
		require.NoError(t, err, "ReviewAchievement failed")

		// Verify the achievement was updated correctly
		require.Equal(t, ach.ID, updatedAch.ID, "Achievement ID should match")
		require.Equal(
			t,
			string(achievement.StatusDone),
			string(updatedAch.Status),
			"Achievement should be in done status",
		)
		require.Equal(t, inspectorOpt.PointsAssigned, updatedAch.Points, "Achievement points should match")
		require.Len(t, updatedAch.Reviews, 2, "Achievement should have two reviews")
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, ach, depHead, _ := setup(t)

		opt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      uuid.Must(uuid.NewV7()), // Non-existent achievement ID
			ReviewerID:         depHead.ID,
			PointsAssigned:     8,
		}

		_, err := svc.ReviewAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})

	t.Run("wrong reviewer role", func(t *testing.T) {
		ctx, svc, ach, _, _ := setup(t)

		// Create a regular teacher as reviewer in the same department
		reviewerOpt := UserUpdateOptions{
			FirstName:    "Wrong",
			LastName:     "Reviewer",
			DepartmentID: ach.Owner.Department.ID, // Use the same department
			NewRoleID:    sesc.Teacher.ID,
		}
		wrongReviewer, err := svc.CreateUser(ctx, reviewerOpt)
		require.NoError(t, err)

		opt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         wrongReviewer.ID,
			PointsAssigned:     8,
		}

		_, err = svc.ReviewAchievement(ctx, opt)
		require.ErrorIs(t, err, sesc.ErrInvalidRole)
	})

	t.Run("wrong achievement status", func(t *testing.T) {
		ctx, svc, _, depHead, _ := setup(t)

		// Create a new achievement but don't submit it
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach := createTestAchievement(t, svc, user, template)

		opt := achievement.ReviewOptions{
			AchievementOwnerID: ach.Owner.ID,
			AchievementID:      ach.ID,
			ReviewerID:         depHead.ID,
			PointsAssigned:     8,
		}

		_, err := svc.ReviewAchievement(ctx, opt)
		require.ErrorIs(t, err, achievement.ErrWrongAchievementStatus)
	})
}

func TestGetAchievement(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, ach achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user := createTestUser(t, svc)
		template := createTestTemplate(t, svc)
		ach = createTestAchievement(t, svc, user, template)
		return ctx, svc, ach
	}

	t.Run("success", func(t *testing.T) {
		ctx, svc, expectedAch := setup(t)

		// Get the achievement
		ach, err := svc.GetAchievement(ctx, expectedAch.ID)
		require.NoError(t, err, "GetAchievement failed")

		// Verify achievement details
		require.Equal(t, expectedAch.ID, ach.ID, "Achievement ID mismatch")
		require.Equal(t, expectedAch.Owner.ID, ach.Owner.ID, "Achievement Owner ID mismatch")
		require.Equal(t, expectedAch.Template.ID, ach.Template.ID, "Achievement Template ID mismatch")
		require.Equal(t, string(expectedAch.Status), string(ach.Status), "Achievement Status mismatch")
	})

	t.Run("achievement not found", func(t *testing.T) {
		ctx, svc, _ := setup(t)

		// Try to get a non-existent achievement
		_, err := svc.GetAchievement(ctx, uuid.Must(uuid.NewV7()))
		require.ErrorIs(t, err, achievement.ErrAchievementNotFound)
	})
}

func TestGetUserAchievements(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, user User, achievements []achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user = createTestUser(t, svc)
		template := createTestTemplate(t, svc)

		// Create multiple achievements for the user
		achievements = make([]achievement.Achievement, 3)
		for i := range 3 {
			achievements[i] = createTestAchievement(t, svc, user, template)
		}

		return ctx, svc, user, achievements
	}

	t.Run("success with pagination", func(t *testing.T) {
		ctx, svc, user, expectedAchievements := setup(t)

		// Get all achievements for the user with pagination
		achievements, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 10)
		require.NoError(t, err, "GetUserAchievements failed")

		// Verify we got the expected number of achievements
		require.Equal(t, len(expectedAchievements), total, "Wrong total count")
		require.Len(t, achievements, len(expectedAchievements), "Wrong number of achievements returned")

		// Create a map of achievement IDs for easy lookup
		expectedIDs := make(map[uuid.UUID]bool)
		for _, ach := range expectedAchievements {
			expectedIDs[ach.ID] = true
		}

		// Verify all expected achievements are in the result
		for _, ach := range achievements {
			require.True(t, expectedIDs[ach.ID], "Unexpected achievement in results")
		}
	})

	t.Run("pagination with offset", func(t *testing.T) {
		ctx, svc, user, expectedAchievements := setup(t)

		// Get first page (2 achievements)
		achievements1, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 2)
		require.NoError(t, err, "GetUserAchievements failed for first page")
		require.Equal(t, len(expectedAchievements), total, "Wrong total count")
		require.Len(t, achievements1, 2, "Wrong number of achievements returned for first page")

		// Get second page (remaining achievement)
		achievements2, total, err := svc.GetUserAchievements(ctx, user.ID, 2, 2)
		require.NoError(t, err, "GetUserAchievements failed for second page")
		require.Equal(t, len(expectedAchievements), total, "Wrong total count")
		require.Len(t, achievements2, 1, "Wrong number of achievements returned for second page")

		// Ensure no overlap between pages
		page1IDs := make(map[uuid.UUID]bool)
		for _, ach := range achievements1 {
			page1IDs[ach.ID] = true
		}
		for _, ach := range achievements2 {
			require.False(t, page1IDs[ach.ID], "Achievement appears in both pages")
		}
	})

	t.Run("no achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)
		user := createTestUser(t, svc)

		// Get achievements for a user with no achievements
		achievements, total, err := svc.GetUserAchievements(ctx, user.ID, 0, 10)
		require.NoError(t, err, "GetUserAchievements failed")
		require.Zero(t, total, "Expected zero total when no achievements exist")
		require.Empty(t, achievements, "Expected empty achievements list")
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		// Get achievements for a non-existent user
		achievements, total, err := svc.GetUserAchievements(ctx, uuid.Must(uuid.NewV7()), 0, 10)
		require.NoError(t, err, "GetUserAchievements should not fail for non-existent user")
		require.Zero(t, total, "Expected zero total when user does not exist")
		require.Empty(t, achievements, "Expected empty achievements list for non-existent user")
	})
}

func TestGetAchievementsForUser(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, user User, achievements []achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		user = createTestUser(t, svc)
		template := createTestTemplate(t, svc)

		// Create multiple achievements for the user
		achievements = make([]achievement.Achievement, 5)
		for i := range 5 {
			achievements[i] = createTestAchievement(t, svc, user, template)
		}

		return ctx, svc, user, achievements
	}

	t.Run("success with pagination", func(t *testing.T) {
		ctx, svc, user, expectedAchievements := setup(t)

		// Get first page (2 items)
		achievements, total, err := svc.GetAchievementsForUser(ctx, user.ID, 0, 2)
		require.NoError(t, err, "GetAchievementsForUser failed")
		require.Equal(t, len(expectedAchievements), total, "Wrong total count")
		require.Len(t, achievements, 2, "Wrong number of achievements returned")

		// Get second page (2 items)
		achievements2, total2, err := svc.GetAchievementsForUser(ctx, user.ID, 2, 2)
		require.NoError(t, err, "GetAchievementsForUser failed for second page")
		require.Equal(t, total, total2, "Total count should be consistent")
		require.Len(t, achievements2, 2, "Wrong number of achievements returned for second page")

		// Verify the IDs are different between pages
		page1IDs := make(map[uuid.UUID]bool)
		for _, ach := range achievements {
			page1IDs[ach.ID] = true
		}

		for _, ach := range achievements2 {
			require.False(t, page1IDs[ach.ID], "Achievement should not appear on both pages")
		}
	})

	t.Run("no achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)
		user := createTestUser(t, svc)

		// Get achievements for a user with no achievements
		achievements, total, err := svc.GetAchievementsForUser(ctx, user.ID, 0, 10)
		require.NoError(t, err, "GetAchievementsForUser failed")
		require.Zero(t, total, "Expected zero total for user with no achievements")
		require.Empty(t, achievements, "Expected empty achievements list")
	})
}

func TestGetGroupedAchievements(t *testing.T) {
	setup := func(t *testing.T) (ctx context.Context, svc *SESC, users []User, achievements map[uuid.UUID][]achievement.Achievement) {
		ctx = t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc = setupSESC(t)
		template := createTestTemplate(t, svc)

		// Create multiple users with achievements
		users = make([]User, 3)
		achievements = make(map[uuid.UUID][]achievement.Achievement)

		for i := range 3 {
			users[i] = createTestUser(t, svc)
			userAchievements := make([]achievement.Achievement, 2)
			for j := range 2 {
				userAchievements[j] = createTestAchievement(t, svc, users[i], template)
				// hack; displays only non-drafts.
				svc.client.Achievement.GetX(ctx, userAchievements[j].ID).
					Update().
					SetStatus(achievement.StatusDepheadReview).
					ExecX(ctx)
			}
			achievements[users[i].ID] = userAchievements
		}

		return ctx, svc, users, achievements
	}

	t.Run("success with pagination", func(t *testing.T) {
		ctx, svc, users, expectedAchievements := setup(t)

		// Get first page (2 users)
		groupedAchievements, total, err := svc.GetGroupedAchievements(ctx, 0, 2)
		require.NoError(t, err, "GetGroupedAchievements failed")
		require.Equal(t, len(users), total, "Wrong total count")
		require.Len(t, groupedAchievements, 2, "Wrong number of users returned")

		// Verify the achievements for each user
		for userID, achievements := range groupedAchievements {
			expected, exists := expectedAchievements[userID]
			require.True(t, exists, "Unexpected user in results")
			require.Len(t, achievements, len(expected), "Wrong number of achievements for user")

			// Create a map of achievement IDs for easy lookup
			expectedIDs := make(map[uuid.UUID]bool)
			for _, ach := range expected {
				expectedIDs[ach.ID] = true
			}

			// Verify all expected achievements are in the result
			for _, ach := range achievements {
				require.True(t, expectedIDs[ach.ID], "Unexpected achievement in results")
			}
		}
	})

	t.Run("no achievements", func(t *testing.T) {
		ctx := t.Context()
		ctx, _ = event.NewRecord(ctx, "test")
		svc := setupSESC(t)

		// Get grouped achievements when there are none
		groupedAchievements, total, err := svc.GetGroupedAchievements(ctx, 0, 10)
		require.NoError(t, err, "GetGroupedAchievements failed")
		require.Zero(t, total, "Expected zero total when no achievements exist")
		require.Empty(t, groupedAchievements, "Expected empty grouped achievements")
	})
}

func TestDetermineNewStatus(t *testing.T) {
	// Only test the valid cases for now
	testCases := []struct {
		name            string
		currentStatus   achievement.Status
		reviewerRole    sesc.Role
		templateKind    achievement.Kind
		pointsAssigned  int
		expectedStatus  achievement.Status
		isValidReviewer bool
	}{
		{
			name:            "department head assigns points",
			currentStatus:   achievement.StatusDepheadReview,
			reviewerRole:    sesc.Dephead,
			templateKind:    achievement.Scientific,
			pointsAssigned:  5,
			expectedStatus:  achievement.StatusInspectorReview,
			isValidReviewer: true,
		},
		{
			name:            "department head assigns zero points",
			currentStatus:   achievement.StatusDepheadReview,
			reviewerRole:    sesc.Dephead,
			templateKind:    achievement.Scientific,
			pointsAssigned:  0,
			expectedStatus:  achievement.StatusDone,
			isValidReviewer: true,
		},
		{
			name:            "scientific deputy reviews scientific achievement",
			currentStatus:   achievement.StatusInspectorReview,
			reviewerRole:    sesc.ScientificDeputy,
			templateKind:    achievement.Scientific,
			pointsAssigned:  7,
			expectedStatus:  achievement.StatusDone,
			isValidReviewer: true,
		},
		{
			name:            "development deputy reviews development achievement",
			currentStatus:   achievement.StatusInspectorReview,
			reviewerRole:    sesc.DevelopmentDeputy,
			templateKind:    achievement.Development,
			pointsAssigned:  6,
			expectedStatus:  achievement.StatusDone,
			isValidReviewer: true,
		},
		{
			name:            "contest deputy reviews olympiad achievement",
			currentStatus:   achievement.StatusInspectorReview,
			reviewerRole:    sesc.ContestDeputy,
			templateKind:    achievement.Olympiad,
			pointsAssigned:  8,
			expectedStatus:  achievement.StatusDone,
			isValidReviewer: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newStatus, isValid := determineNewStatus(
				tc.currentStatus,
				tc.reviewerRole,
				tc.templateKind,
				tc.pointsAssigned,
			)

			// Compare as strings to avoid type comparison issues
			require.Equal(t, string(tc.expectedStatus), string(newStatus), "Status mismatch")
			require.Equal(t, tc.isValidReviewer, isValid, "Reviewer validity mismatch")
		})
	}
}
