package db_local

import "testing"

func TestClassifyPgMigration(t *testing.T) {
	cases := []struct {
		name       string
		oldVersion string
		target     string
		want       pgMigrationKind
	}{
		{
			name:       "same-major older minor (14.17 -> 14.19) is minor",
			oldVersion: "14.17.0",
			target:     "14.19.0",
			want:       pgMigrationMinor,
		},
		{
			name:       "same-major newer minor still minor",
			oldVersion: "14.19.0",
			target:     "14.17.0",
			want:       pgMigrationMinor,
		},
		{
			name:       "identical version is minor",
			oldVersion: "14.19.0",
			target:     "14.19.0",
			want:       pgMigrationMinor,
		},
		{
			name:       "cross-major (14 -> 18) is major",
			oldVersion: "14.19.0",
			target:     "18.0.0",
			want:       pgMigrationMajor,
		},
		{
			name:       "cross-major different minors (14.2 -> 18.4) is major",
			oldVersion: "14.2.0",
			target:     "18.4.1",
			want:       pgMigrationMajor,
		},
		{
			name:       "cross-major downgrade (18 -> 14) is major",
			oldVersion: "18.0.0",
			target:     "14.19.0",
			want:       pgMigrationMajor,
		},
		{
			name:       "unparseable old version falls back to minor (preserve auto restore)",
			oldVersion: "not-a-version",
			target:     "14.19.0",
			want:       pgMigrationMinor,
		},
		{
			name:       "empty target falls back to minor",
			oldVersion: "14.19.0",
			target:     "",
			want:       pgMigrationMinor,
		},
		{
			name:       "both unparseable falls back to minor",
			oldVersion: "foo",
			target:     "bar",
			want:       pgMigrationMinor,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyPgMigration(tc.oldVersion, tc.target)
			if got != tc.want {
				t.Errorf("classifyPgMigration(%q, %q) = %d, want %d", tc.oldVersion, tc.target, got, tc.want)
			}
		})
	}
}
