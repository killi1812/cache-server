package model

// Agent represents a deployment agent belonging to a workspace.
type Agent struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"type:varchar(100);unique;not null" json:"name"`

	WorkspaceId uint       `json:"workspace_id"`
	Workspace   *Workspace `gorm:"foreignKey:WorkspaceId;constraint:OnUpdate:CASCADE;" json:"-"`
}
