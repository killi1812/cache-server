// Package model hosts all models for database
package model

// GetAllModels returns an array of all models
func GetAllModels() []any {
	return []any{
		&BinaryCache{}, // 1. Parent of StorePath and Workspace
		&Workspace{},   // 2. Child of Cache, Parent of Agent
		&Agent{},       // 3. Child of Workspace
		&StorePath{},   // 4. Child of Cache
		&Deployment{},  // 5. Child of Agent
	}
}
