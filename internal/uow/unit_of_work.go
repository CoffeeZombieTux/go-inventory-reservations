package uow

import (
	"context"
	"database/sql"
)

// UnitOfWork represents a unit of work.
type UnitOfWork struct {
	tx *sql.Tx
}

// UnitOfWorkManager manages the unit of work.
type UnitOfWorkManager struct {
	db *sql.DB
}

// NewUnitOfWorkManager creates a new UnitOfWorkManager instance.
func NewUnitOfWorkManager(db *sql.DB) *UnitOfWorkManager {
	return &UnitOfWorkManager{db: db}
}

// Begin starts a new unit of work.
func (u *UnitOfWorkManager) Begin(ctx context.Context) (*UnitOfWork, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &UnitOfWork{
		tx: tx,
	}, nil
}

// GetTransaction returns the transaction associated with the unit of work.
func (u *UnitOfWork) GetTransaction() (tx *sql.Tx) {
	return u.tx
}

// Commit the unit of work
func (u *UnitOfWork) Commit() error {
	return u.tx.Commit()
}

// Rollback the unit of work
func (u *UnitOfWork) Rollback() error {
	return u.tx.Rollback()
}
