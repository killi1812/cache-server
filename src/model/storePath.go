package model

import "gorm.io/gorm"

// StorePath represents a specific store path entry inside a binary cache.
type StorePath struct {
	gorm.Model
	StoreHash   string `gorm:"size:255;not null" json:"storeHash"`
	StoreSuffix string `gorm:"size:2048;not null" json:"storeSuffix"`
	FileHash    string `gorm:"size:255" json:"fileHash"`
	FileSize    int64  `json:"fileSize"`
	NarHash     string `gorm:"size:255" json:"narHash"`
	NarSize     int64  `json:"narSize"`
	Deriver     string `gorm:"size:4096" json:"deriver"`
	References  string `gorm:"size:4096" json:"references"`

	BinaryCacheId uint         `gorm:"column:binary_cache_id" json:"binaryCacheId"`
	BinaryCache   *BinaryCache `gorm:"foreignKey:BinaryCacheId" json:"-"`
}
