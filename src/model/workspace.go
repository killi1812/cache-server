package model

import "gorm.io/gorm"

type Workspace struct {
	gorm.Model
	Name   string  `gorm:"type:varchar(100);unique;not null"`
	Token  string  `gorm:"type:varchar(255);not null"`
	Agents []Agent `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"agents,omitempty"`
}
