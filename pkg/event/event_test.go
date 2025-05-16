package event_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kozlov-ma/sesc-backend/pkg/event"
	"github.com/stretchr/testify/require"
)

func TestRecord_Add(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		_, rec := event.NewRecord(t.Context(), "test")

		rec.Add("type", "test_event", "message", "it_works")

		vals := rec.AllValues()

		require.Equal(t, "test_event", vals["type"])
		require.Equal(t, "it_works", vals["message"])
	})
	t.Run("sub_add", func(t *testing.T) {
		_, rec := event.NewRecord(t.Context(), "test")

		rec.Add("type", "test_event", "message", "it_works")

		rec.Sub("stage_1").Add(
			"stage", 1,
			"works", true,
		)

		rec.Sub("stage_2").Add(
			"wide_events", "are cool",
		)

		vals := rec.AllValues()

		require.Equal(t, "test_event", vals["type"])
		require.Equal(t, "it_works", vals["message"])

		require.Equal(t, 1, vals["stage_1.stage"])
		require.Equal(t, true, vals["stage_1.works"])

		require.Equal(t, "are cool", vals["stage_2.wide_events"])
	})
	t.Run("characters", func(t *testing.T) {
		_, rec := event.NewRecord(t.Context(), "test")

		rec.Add("@flag", "a", "_underscore", "__")
	})
	t.Run("errors", func(t *testing.T) {
		_, rec := event.NewRecord(t.Context(), "test")

		e1 := errors.New("e1")
		e2 := errors.New("e2")
		e3 := fmt.Errorf("e3: %w", e2)
		rec.Add("error", e1)
		rec.Add("error", e2)
		rec.Add("error", e3)

		require.ErrorIs(t, rec.Value("error").(error), e1)
		require.ErrorIs(t, rec.Value("error").(error), e2)
		require.ErrorIs(t, rec.Value("error").(error), e3)
	})
}

func TestRecord_Value(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		_, rec := event.NewRecord(t.Context(), "test")

		rec.Add("type", "test_event", "message", "it_works")

		require.Equal(t, "test_event", rec.Value("type"))
		require.Equal(t, "it_works", rec.Value("message"))
	})
	t.Run("subrecord", func(t *testing.T) {
		t.Run("sub_add", func(t *testing.T) {
			_, rec := event.NewRecord(t.Context(), "test")
			rec.Add("type", "test_event", "message", "it_works")

			rec.Sub("stage_1").Add(
				"stage", 1,
				"works", true,
			)

			rec.Sub("stage_2").Add(
				"wide_events", "are cool",
			)

			require.Equal(t, "test_event", rec.Value("type"))
			require.Equal(t, "it_works", rec.Value("message"))

			require.Equal(t, 1, rec.Value("stage_1.stage"))
			require.Equal(t, true, rec.Value("stage_1.works"))

			require.Equal(t, "are cool", rec.Value("stage_2.wide_events"))
		})
	})
}

func TestContextOperations(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		ctx, rec := event.NewRecord(t.Context(), "test")

		func(ctx context.Context) {
			rec := event.Get(ctx)
			rec.Sub("group").Set(
				"k1", "v1",
				"k2", "v2",
			)
		}(ctx)

		vals := rec.AllValues()
		expected := map[string]any{
			"group.k1": "v1",
			"group.k2": "v2",
		}

		require.Subset(t, vals, expected)
	})
}
