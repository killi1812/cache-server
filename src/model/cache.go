package model

// BinaryCache represents a cache that stores multiple paths and belongs to multiple workspaces.
type BinaryCache struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"type:varchar(100);unique;not null" json:"name"`
	URL       string `gorm:"type:varchar(255);not null" json:"url"`
	Token     string `gorm:"type:varchar(255);not null" json:"-"` // not serialized
	Access    string `gorm:"type:varchar(50)" json:"access"`
	Port      int    `gorm:"unique" json:"port"` // TODO: check if port needs to be unique
	Retention int    `json:"retention"`

	StorePaths []StorePath `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:CASCADE;" json:"store_paths,omitempty"`
	Workspaces []Workspace `gorm:"foreignKey:BinaryCacheId;constraint:OnDelete:SET NULL;" json:"workspaces,omitempty"`
}
