package unit_test

import (
	"testing"

	"go-gin-clean/internal/infrastructure/repository/postgres"
)

var testUserSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"email":      "email",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func TestSanitizeSortColumn(t *testing.T) {
	t.Run("AC-009: Allowed columns return exact column and sort order", func(t *testing.T) {
		col, dir := postgres.SanitizeSortColumn("name", "asc", testUserSortColumns)
		if col != "name" || dir != "asc" {
			t.Fatalf("expected ('name', 'asc'), got ('%s', '%s')", col, dir)
		}

		col, dir = postgres.SanitizeSortColumn("email", "desc", testUserSortColumns)
		if col != "email" || dir != "desc" {
			t.Fatalf("expected ('email', 'desc'), got ('%s', '%s')", col, dir)
		}
	})

	t.Run("AC-009: Malicious SQL injection in sortBy is sanitized to default created_at", func(t *testing.T) {
		injections := []string{
			"name; DROP TABLE users;--",
			"id OR 1=1",
			"(CASE WHEN (SELECT 1=1) THEN name ELSE email END)",
			"sleep(5)",
			"UNION SELECT * FROM users",
		}

		for _, attack := range injections {
			col, dir := postgres.SanitizeSortColumn(attack, "desc", testUserSortColumns)
			if col != "created_at" {
				t.Fatalf("SQL injection vulnerability: unapproved column %q was not sanitized! got: %q", attack, col)
			}
			if dir != "desc" {
				t.Fatalf("unexpected dir: %q", dir)
			}
		}
	})

	t.Run("AC-009: Unknown sort direction defaults to desc", func(t *testing.T) {
		col, dir := postgres.SanitizeSortColumn("name", "invalid; DROP DATABASE;", testUserSortColumns)
		if col != "name" || dir != "desc" {
			t.Fatalf("expected ('name', 'desc'), got ('%s', '%s')", col, dir)
		}
	})
}
