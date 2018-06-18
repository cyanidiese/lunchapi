package models

import "time"

type Order struct {
	Id           int64      `gorm:"column:id" gorm:"primary_key" json:"id"`
	UserId       int64      `gorm:"column:user_id" json:"user_id"`
	ItemId       int64      `gorm:"column:item_id" json:"itemId"`
	OrderedCount int64      `gorm:"column:ordered_count" json:"orderedCount"`
	MenuItem     MenuItem   `gorm:"foreignkey:ItemId;association_foreignkey:Id" json:"item"`
	CreatedAt    time.Time  `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt    time.Time  `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Order) TableName() string {
	return "orders"
}
