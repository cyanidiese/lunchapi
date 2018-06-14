package models

import "time"

type MenuItem struct {
	Id             int64     `gorm:"column:id" gorm:"primary_key" json:"id"`
	MenuId         int64     `gorm:"column:menu_id" json:"-"`
	DishId         int64     `gorm:"column:dish_id" json:"dishId"`
	InitialCount   int64     `gorm:"column:initial_count" json:"initialCount"`
	AvailableCount int64     `gorm:"column:available_count" json:"availableCount"`
	Price          float64   `gorm:"column:price" json:"price"`
	CreatedAt      time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt      time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (MenuItem) TableName() string {
	return "menu_items"
}
