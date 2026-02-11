// Package model hosts all models for database
package model

// GetAllModels returns an array of all models
func GetAllModels() []any {
	return []any{
		&Workspace{},
		&Agent{},
	}
}
