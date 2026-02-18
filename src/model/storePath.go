package model

// StorePath represents a specific store path entry inside a binary cache.
type StorePath struct {
	ID          uint   `gorm:"primarykey"`
	StoreHash   string `gorm:"type:varchar(255);not null" json:"storeHash"`
	StoreSuffix string `gorm:"type:varchar(255);not null" json:"storeSuffix"`
	FileHash    string `gorm:"type:varchar(255)" json:"fileHash"`
	FileSize    int64  `json:"fileSize"`
	NarHash     string `gorm:"type:varchar(255)" json:"narHash"`
	NarSize     int64  `json:"narSize"`
	Deriver     string `gorm:"type:varchar(255)" json:"deriver"`
	References  string `gorm:"type:text" json:"references"` // Stored as a string (serialized)

	BinaryCacheId uint         `json:"binary_cache_id"`
	BinaryCache   *BinaryCache `gorm:"foreignKey:BinaryCacheId" json:"-"`
}
