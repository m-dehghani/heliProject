package entity

type Account struct {
	ID         uint `gorm:"primaryKey"`
	CustomerID uint
	Balance    float64
}
