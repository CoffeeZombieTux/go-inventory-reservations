package apperror

import (
	"database/sql"
	"errors"
	"testing"
)

func TestNewWrapAsError(t *testing.T) {
	err := New(CodeValidationError, "bad input", "field x")
	appErr, ok := As(err)
	if !ok {
		t.Fatalf("expected structured error")
	}
	if appErr.Code != CodeValidationError || appErr.Message != "bad input" {
		t.Fatalf("unexpected app error: %+v", appErr)
	}

	cause := errors.New("root")
	wrapped := Wrap(cause, CodeInternalError, "internal")
	if wrapped == nil {
		t.Fatalf("expected wrapped error")
	}
	if !errors.Is(wrapped, cause) {
		t.Fatalf("expected wrapped to contain cause")
	}
}

func TestFromDBMappings(t *testing.T) {
	err := FromDB(sql.ErrNoRows, "", "")
	appErr, _ := As(err)
	if appErr.Code != CodeDBNoRows {
		t.Fatalf("expected DB no rows code, got %s", appErr.Code)
	}

	err = FromDB(errors.New("duplicate key value violates unique constraint"), "", "")
	appErr, _ = As(err)
	if appErr.Code != CodeDBDuplicateKey {
		t.Fatalf("expected duplicate key code, got %s", appErr.Code)
	}

	err = FromDB(errors.New("violates foreign key constraint"), "", "")
	appErr, _ = As(err)
	if appErr.Code != CodeDBForeignKeyConstraint {
		t.Fatalf("expected fk code, got %s", appErr.Code)
	}

	err = FromDB(errors.New("violates check constraint"), "", "")
	appErr, _ = As(err)
	if appErr.Code != CodeDBCheckConstraint {
		t.Fatalf("expected check code, got %s", appErr.Code)
	}

	err = FromDB(errors.New("unknown"), "fallback msg", "FALLBACK")
	appErr, _ = As(err)
	if appErr.Code != "FALLBACK" || appErr.Message != "fallback msg" {
		t.Fatalf("unexpected fallback mapping: %+v", appErr)
	}
}
