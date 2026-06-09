package main

import "testing"

func TestSQLiteDatabasePath(t *testing.T) {
	tests := []struct {
		name           string
		dataSourceName string
		want           string
	}{
		{
			name:           "memory database",
			dataSourceName: ":memory:",
			want:           "",
		},
		{
			name:           "plain relative path",
			dataSourceName: "./data/app.db",
			want:           "./data/app.db",
		},
		{
			name:           "plain absolute path",
			dataSourceName: "/tmp/comelymd/comelymd.db",
			want:           "/tmp/comelymd/comelymd.db",
		},
		{
			name:           "file URI relative path",
			dataSourceName: "file:./data/app.db?cache=shared",
			want:           "./data/app.db",
		},
		{
			name:           "file URI absolute path",
			dataSourceName: "file:/tmp/comelymd.db?mode=rwc",
			want:           "/tmp/comelymd.db",
		},
		{
			name:           "file URI absolute path with authority-style slashes",
			dataSourceName: "file:///tmp/comelymd.db?mode=rwc",
			want:           "/tmp/comelymd.db",
		},
		{
			name:           "file URI memory database",
			dataSourceName: "file::memory:?cache=shared",
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqliteDatabasePath(tt.dataSourceName); got != tt.want {
				t.Fatalf("sqliteDatabasePath(%q) = %q, want %q", tt.dataSourceName, got, tt.want)
			}
		})
	}
}
