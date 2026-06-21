// Package domain provides domain functionality for the IMB platform.
package domain

// Institute Model
type Institution struct {
	Base

	Name     string `gorm:"type:varchar(255);not null"`
	Code     string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Address  string `gorm:"type:text"`
	Phone    string `gorm:"type:varchar(20)"`
	Email    string `gorm:"type:varchar(255)"`
	IsActive bool   `gorm:"default:true"`
}
