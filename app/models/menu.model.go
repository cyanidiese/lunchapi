package models

import "time"

type Menu struct {
	Id           int64      `gorm:"column:id" gorm:"primary_key" json:"id"`
	ProviderId   int64      `gorm:"column:provider_id" json:"-"`
	Date         string     `gorm:"column:date" sql:"type:date" json:"date"`
	DeliveryTime string     `gorm:"column:time" sql:"type:time" json:"time"`
	DeadlineAt   string     `gorm:"column:deadline_at" sql:"type:datetime" json:"deadline"`
	Items        []MenuItem `gorm:"foreignkey:MenuId;association_foreignkey:Id" json:"items"`
	CreatedAt    time.Time  `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt    time.Time  `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Menu) TableName() string {
	return "menus"
}
