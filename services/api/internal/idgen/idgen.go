// Package idgen sinh UUID v7 (time-ordered) cho mọi entity.
package idgen

import "github.com/google/uuid"

// New trả về uuid v7 dạng string. Time-ordered → tốt cho B-tree index.
func New() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Cực hiếm; fallback uuid v4.
		return uuid.NewString()
	}
	return id.String()
}

// NewUUID trả về uuid.UUID (cho pgx tham số trực tiếp).
func NewUUID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
