package model

import "gorm.io/gorm"

type Agent struct {
	gorm.Model
	Name        string `gorm:"type:varchar(100);unique;not null"`
	token       string `gorm:"type:varchar(255);not null"`
	WorkspaceId uint
	Workspace   *Workspace `gorm:"foreignKey:WorkspaceId;constraint:OnUpdate:CASCADE;" json:"-"`
}
