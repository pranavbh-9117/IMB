package seed

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/pranavbh-9117/IMB/internal/domain"
	"github.com/pranavbh-9117/IMB/pkg/config"
	"github.com/pranavbh-9117/IMB/pkg/password"
)

// Run seeds the Super Admin account if one does not already exist.
// It is safe to call on every application startup — if a user with the
// configured email is found the function returns nil immediately without
// modifying any data.
//
// The caller is responsible for providing an open, healthy *gorm.DB instance
// and a fully populated config.SeedConfig.
func Run(db *gorm.DB, cfg config.SeedConfig) error {
	var existing domain.User

	err := db.Where("email = ?", cfg.SuperAdminEmail).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("seed: failed to query for existing super admin: %w", err)
	}

	if err == nil {
		// Super Admin already exists — nothing to do.
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
