package entity

import (
	"time"
)

type Transaction struct {
	ID         uint `gorm:"primaryKey"`
	CustomerID uint
	Type       string
	Amount     float64
	Date       time.Time
}
