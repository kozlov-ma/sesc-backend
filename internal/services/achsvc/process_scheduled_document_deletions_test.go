package achsvc

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/achievement"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/enttest"
	"github.com/kozlov-ma/sesc-backend/db/entdb/ent/file"
	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStorage struct {
	removedObjects []string
}

func (m *mockStorage) RemoveObject(_ context.Context, objectKey string) error {
	m.removedObjects = append(m.removedObjects, objectKey)
	return nil
}

func TestProcessScheduledDocumentDeletions(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	ctx, _ := event.NewRecord(t.Context(), "test")
	service := &ACS{client: client}

	t.Run("deletes scheduled documents past cutoff", func(t *testing.T) {
		storage := &mockStorage{}

		user := client.User.Create().
			SetID(uuid.Must(uuid.NewV4())).
			SetFirstName("Test").
			SetLastName("User").
			SetEmploymentRate(1.0).
			SetEmploymentType(1).
			SetPersonnelCategory(1).
			SetRole(1).
			SetJobTitle("Dev").
			SetSubdivision("IT").
			SaveX(ctx)

		group := client.AchievementGroup.Create().
			SetName("Group1").
			SetDescription("Test").
			SaveX(ctx)

		template := client.AchievementTemplate.Create().
			SetName("Template1").
			SetDescription("Test").
			SetPointsLimit(100).
			SetReviewerRole(1).
			SetGroup(group).
			SaveX(ctx)

		ach := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		fileRec := client.File.Create().
			SetName("test.pdf").
			SetS3ObjectKey("key-1").
			SetSize(1024).
			SetOwner(user).
			SaveX(ctx)

		oldTime := time.Now().Add(-48 * time.Hour)
		doc := client.AchievementDocument.Create().
			SetAchievement(ach).
			SetName("Doc1").
			SetFile(fileRec).
			SetStatus(achievement.DocumentStatusScheduled).
			SetScheduledDeletionAt(oldTime).
			SaveX(ctx)

		err := service.ProcessScheduledDocumentDeletions(ctx, storage, 24*time.Hour)
		require.NoError(t, err)

		updatedDoc, err := client.AchievementDocument.Get(ctx, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, achievement.DocumentStatusDeleted, updatedDoc.Status)
		assert.Nil(t, updatedDoc.FileID)

		fileExists, err := client.File.Query().Where(file.ID(fileRec.ID)).Exist(ctx)
		require.NoError(t, err)
		assert.False(t, fileExists)
		assert.Contains(t, storage.removedObjects, "key-1")
	})

	t.Run("keeps file when referenced by active document", func(t *testing.T) {
		storage := &mockStorage{}

		user := client.User.Create().
			SetID(uuid.Must(uuid.NewV4())).
			SetFirstName("Test").
			SetLastName("User").
			SetEmploymentRate(1.0).
			SetEmploymentType(1).
			SetPersonnelCategory(1).
			SetRole(1).
			SetJobTitle("Dev").
			SetSubdivision("IT").
			SaveX(ctx)

		group := client.AchievementGroup.Create().
			SetName("Group2").
			SetDescription("Test").
			SaveX(ctx)

		template := client.AchievementTemplate.Create().
			SetName("Template2").
			SetDescription("Test").
			SetPointsLimit(100).
			SetReviewerRole(1).
			SetGroup(group).
			SaveX(ctx)

		ach1 := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		ach2 := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		sharedFile := client.File.Create().
			SetName("shared.pdf").
			SetS3ObjectKey("shared-key").
			SetSize(2048).
			SetOwner(user).
			SaveX(ctx)

		oldTime := time.Now().Add(-48 * time.Hour)
		doc1 := client.AchievementDocument.Create().
			SetAchievement(ach1).
			SetName("Doc1").
			SetFile(sharedFile).
			SetStatus(achievement.DocumentStatusScheduled).
			SetScheduledDeletionAt(oldTime).
			SaveX(ctx)

		client.AchievementDocument.Create().
			SetAchievement(ach2).
			SetName("Doc2").
			SetFile(sharedFile).
			SetStatus(achievement.DocumentStatusActive).
			SaveX(ctx)

		err := service.ProcessScheduledDocumentDeletions(ctx, storage, 24*time.Hour)
		require.NoError(t, err)

		updatedDoc1, err := client.AchievementDocument.Get(ctx, doc1.ID)
		require.NoError(t, err)
		assert.Equal(t, achievement.DocumentStatusDeleted, updatedDoc1.Status)

		fileExists, err := client.File.Query().Where(file.ID(sharedFile.ID)).Exist(ctx)
		require.NoError(t, err)
		assert.True(t, fileExists)
		assert.NotContains(t, storage.removedObjects, "shared-key")
	})

	t.Run("does not delete documents not past cutoff", func(t *testing.T) {
		storage := &mockStorage{}

		user := client.User.Create().
			SetID(uuid.Must(uuid.NewV4())).
			SetFirstName("Test").
			SetLastName("User").
			SetEmploymentRate(1.0).
			SetEmploymentType(1).
			SetPersonnelCategory(1).
			SetRole(1).
			SetJobTitle("Dev").
			SetSubdivision("IT").
			SaveX(ctx)

		group := client.AchievementGroup.Create().
			SetName("Group3").
			SetDescription("Test").
			SaveX(ctx)

		template := client.AchievementTemplate.Create().
			SetName("Template3").
			SetDescription("Test").
			SetPointsLimit(100).
			SetReviewerRole(1).
			SetGroup(group).
			SaveX(ctx)

		ach := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		fileRec := client.File.Create().
			SetName("recent.pdf").
			SetS3ObjectKey("recent-key").
			SetSize(512).
			SetOwner(user).
			SaveX(ctx)

		recentTime := time.Now().Add(-12 * time.Hour)
		doc := client.AchievementDocument.Create().
			SetAchievement(ach).
			SetName("RecentDoc").
			SetFile(fileRec).
			SetStatus(achievement.DocumentStatusScheduled).
			SetScheduledDeletionAt(recentTime).
			SaveX(ctx)

		err := service.ProcessScheduledDocumentDeletions(ctx, storage, 24*time.Hour)
		require.NoError(t, err)

		updatedDoc, err := client.AchievementDocument.Get(ctx, doc.ID)
		require.NoError(t, err)
		assert.Equal(t, achievement.DocumentStatusScheduled, updatedDoc.Status)

		fileExists, err := client.File.Query().Where(file.ID(fileRec.ID)).Exist(ctx)
		require.NoError(t, err)
		assert.True(t, fileExists)
		assert.NotContains(t, storage.removedObjects, "recent-key")
	})

	t.Run("deletes file when all referencing documents are scheduled", func(t *testing.T) {
		storage := &mockStorage{}

		user := client.User.Create().
			SetID(uuid.Must(uuid.NewV4())).
			SetFirstName("Test").
			SetLastName("User").
			SetEmploymentRate(1.0).
			SetEmploymentType(1).
			SetPersonnelCategory(1).
			SetRole(1).
			SetJobTitle("Dev").
			SetSubdivision("IT").
			SaveX(ctx)

		group := client.AchievementGroup.Create().
			SetName("Group4").
			SetDescription("Test").
			SaveX(ctx)

		template := client.AchievementTemplate.Create().
			SetName("Template4").
			SetDescription("Test").
			SetPointsLimit(100).
			SetReviewerRole(1).
			SetGroup(group).
			SaveX(ctx)

		ach1 := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		ach2 := client.Achievement.Create().
			SetOwner(user).
			SetTemplate(template).
			SetPoints(0).
			SetStatus(achievement.StatusDraft).
			SaveX(ctx)

		sharedFile := client.File.Create().
			SetName("shared2.pdf").
			SetS3ObjectKey("shared-key-2").
			SetSize(3072).
			SetOwner(user).
			SaveX(ctx)

		oldTime := time.Now().Add(-48 * time.Hour)
		doc1 := client.AchievementDocument.Create().
			SetAchievement(ach1).
			SetName("DocA").
			SetFile(sharedFile).
			SetStatus(achievement.DocumentStatusScheduled).
			SetScheduledDeletionAt(oldTime).
			SaveX(ctx)

		doc2 := client.AchievementDocument.Create().
			SetAchievement(ach2).
			SetName("DocB").
			SetFile(sharedFile).
			SetStatus(achievement.DocumentStatusScheduled).
			SetScheduledDeletionAt(oldTime).
			SaveX(ctx)

		err := service.ProcessScheduledDocumentDeletions(ctx, storage, 24*time.Hour)
		require.NoError(t, err)

		updatedDoc1, err := client.AchievementDocument.Get(ctx, doc1.ID)
		require.NoError(t, err)
		assert.Equal(t, achievement.DocumentStatusDeleted, updatedDoc1.Status)
		assert.Nil(t, updatedDoc1.FileID)

		updatedDoc2, err := client.AchievementDocument.Get(ctx, doc2.ID)
		require.NoError(t, err)
		assert.Equal(t, achievement.DocumentStatusDeleted, updatedDoc2.Status)
		assert.Nil(t, updatedDoc2.FileID)

		fileExists, err := client.File.Query().Where(file.ID(sharedFile.ID)).Exist(ctx)
		require.NoError(t, err)
		assert.False(t, fileExists)
		assert.Contains(t, storage.removedObjects, "shared-key-2")
	})
}
