package sesc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAchievementKind_String(t *testing.T) {
	tests := []struct {
		name string
		kind AchievementKind
		want string
	}{
		{
			name: "olympiad",
			kind: Olympiad,
			want: "olympiad",
		},
		{
			name: "development",
			kind: Development,
			want: "development",
		},
		{
			name: "scientific",
			kind: Scientific,
			want: "scientific",
		},
		{
			name: "invalid",
			kind: AchievementKind("invalid"),
			want: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.kind.String())
		})
	}
}

func TestAchievementKind_IsValid(t *testing.T) {
	tests := []struct {
		name string
		kind AchievementKind
		want bool
	}{
		{
			name: "olympiad_valid",
			kind: Olympiad,
			want: true,
		},
		{
			name: "development_valid",
			kind: Development,
			want: true,
		},
		{
			name: "scientific_valid",
			kind: Scientific,
			want: true,
		},
		{
			name: "empty_invalid",
			kind: AchievementKind(""),
			want: false,
		},
		{
			name: "random_invalid",
			kind: AchievementKind("random"),
			want: false,
		},
		{
			name: "uppercase_invalid",
			kind: AchievementKind("OLYMPIAD"),
			want: false,
		},
		{
			name: "mixed_case_invalid",
			kind: AchievementKind("Olympiad"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.kind.IsValid())
		})
	}
}

func TestAchievementKind_Validate(t *testing.T) {
	tests := []struct {
		name    string
		kind    AchievementKind
		wantErr error
	}{
		{
			name:    "olympiad_valid",
			kind:    Olympiad,
			wantErr: nil,
		},
		{
			name:    "development_valid",
			kind:    Development,
			wantErr: nil,
		},
		{
			name:    "scientific_valid",
			kind:    Scientific,
			wantErr: nil,
		},
		{
			name:    "empty_invalid",
			kind:    AchievementKind(""),
			wantErr: ErrInvalidAchievementKind,
		},
		{
			name:    "random_invalid",
			kind:    AchievementKind("random"),
			wantErr: ErrInvalidAchievementKind,
		},
		{
			name:    "uppercase_invalid",
			kind:    AchievementKind("OLYMPIAD"),
			wantErr: ErrInvalidAchievementKind,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.kind.Validate()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
