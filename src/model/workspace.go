package model

// Workspace represents a deployment environment linked to a specific binary cache.
type Workspace struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"type:varchar(100);unique;not null" json:"name"`
	Token string `gorm:"type:varchar(255);not null" json:"-"` // not serialized deploy token

	BinaryCacheId uint         `json:"binary_cache_id"`
	BinaryCache   *BinaryCache `gorm:"foreignKey:BinaryCacheId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"binary_cache,omitempty"`

	Agents []Agent `gorm:"foreignKey:WorkspaceId;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"agents,omitempty"`
}
