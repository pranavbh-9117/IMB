package migration

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// Run executes GORM AutoMigrate for all domain entities in dependency order.
// Tables are created or altered to match the current struct definitions.
// Existing columns and data are never dropped or truncated.
//
// The caller is responsible for providing an open, healthy *gorm.DB instance.
// Run should be called once at application startup, before the HTTP server
// begins accepting requests.
func Run(db *gorm.DB) error {
	err := db.AutoMigrate(
		// Tier 1 — no foreign key dependencies
		&domain.Institution{},

		// Tier 2 — depends on Institution
		&domain.User{},

		// Tier 3 — depends on User
		&domain.RefreshToken{},

		// Tier 4 — depends on User + Institution
		&domain.LeaveBalance{},
		&domain.LeaveRequest{},

		// Tier 5 — depends on Institution + User
		&domain.Quiz{},

		// Tier 6 — depends on Quiz
		&domain.Question{},

		// Tier 7 — depends on Question
		&domain.Option{},

		// Tier 8 — depends on Quiz + User
		&domain.QuizAttempt{},

		// Tier 9 — depends on QuizAttempt + Question + Option
		&domain.QuizAnswer{},
	)
	if err != nil {
		return fmt.Errorf("migration: auto-migrate failed: %w", err)
	}

	return nil
}
