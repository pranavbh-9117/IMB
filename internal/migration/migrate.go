// Package migration provides migration functionality for the IMB platform.
package migration

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
)

// Migrate the Go struct models into DB Tables
func Run(db *gorm.DB) error {
	err := db.AutoMigrate(

		&domain.Institution{},
		&domain.User{},
		&domain.RefreshToken{},
		&domain.LeaveBalance{},
		&domain.LeaveRequest{},
		&domain.Quiz{},
		&domain.Question{},
		&domain.Option{},
		&domain.QuizAttempt{},
		&domain.QuizAnswer{},
	)
	if err != nil {
		return fmt.Errorf("migration: auto-migrate failed: %w", err)
	}

	return nil
}
