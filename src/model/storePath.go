package model

import "gorm.io/gorm"

// StorePath represents a specific store path entry inside a binary cache.
type StorePath struct {
	gorm.Model
	StoreHash   string `gorm:"type:text;not null" json:"storeHash"`
	StoreSuffix string `gorm:"type:text;not null" json:"storeSuffix"`
	FileHash    string `gorm:"type:text" json:"fileHash"`
	FileSize    int64  `json:"fileSize"`
	NarHash     string `gorm:"type:text" json:"narHash"`
	NarSize     int64  `json:"narSize"`
	Deriver     string `gorm:"type:text" json:"deriver"`
	References  string `gorm:"type:text" json:"references"`

	// BinaryCacheId uint         `json:"binarycacheId"`
	// BinaryCache   *BinaryCache `json:"-"`
	BinaryCacheId uint         `gorm:"column:binary_cache_id" json:"binaryCacheId"`
	BinaryCache   *BinaryCache `gorm:"foreignKey:BinaryCacheId" json:"-"`
}
