// Package seed provides seed functionality for the IMB platform.
package seed

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/password"
)

// Populate Super Admin to the DB.
func Run(db *gorm.DB, cfg config.SeedConfig) error {
	var existing domain.User

	err := db.Where("email = ?", cfg.SuperAdminEmail).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("seed: failed to query for existing super admin: %w", err)
	}

	if err == nil {
		return nil
	}

	hash, err := password.Hash(cfg.SuperAdminPassword)
	if err != nil {
		return fmt.Errorf("seed: failed to hash super admin password: %w", err)
	}

	admin := domain.User{
		Name:          "Super Admin",
		Email:         cfg.SuperAdminEmail,
		PasswordHash:  hash,
		Role:          domain.RoleSuperAdmin,
		IsActive:      true,
		InstitutionID: nil,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed: failed to create super admin: %w", err)
	}

	return nil
}
