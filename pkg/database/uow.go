// Package database provides infrastructure persistence abstractions and session management.
package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionMetaKey struct{}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]*gorm.DB)
)

// UnitOfWork defines the transactional contract for services without exposing ORM internals.
type UnitOfWork interface {
	WithinReadOnlyTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormUnitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork creates a new UnitOfWork encapsulated within the infrastructure layer.
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &gormUnitOfWork{db: db}
}

// WithinReadOnlyTransaction executes fn inside a read-only database transaction.
// Transaction management remains completely encapsulated within the persistence implementation.
func (u *gormUnitOfWork) WithinReadOnlyTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx := u.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("uow: begin transaction failed: %w", err)
	}

	// Try setting transaction read-only where supported by driver
	tx.Exec("SET TRANSACTION READ ONLY")

	sessionID := uuid.NewString()
	registryMu.Lock()
	registry[sessionID] = tx
	registryMu.Unlock()

	defer func() {
		registryMu.Lock()
		delete(registry, sessionID)
		registryMu.Unlock()

		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	txCtx := context.WithValue(ctx, sessionMetaKey{}, sessionID)

	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("uow: commit failed: %w", err)
	}
	return nil
}

// GetSession returns the active session encapsulated within the persistence layer registry.
// This allows repositories to participate in the active UnitOfWork transaction without inspecting context for ORM objects.
func GetSession(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if id, ok := ctx.Value(sessionMetaKey{}).(string); ok && id != "" {
		registryMu.RLock()
		tx, found := registry[id]
		registryMu.RUnlock()
		if found && tx != nil {
			return tx.WithContext(ctx)
		}
	}
	return defaultDB.WithContext(ctx)
}
