package models

import "time"

type Image struct {
	Id        int64    `gorm:"column:id" gorm:"primary_key" json:"id"`
	UserId    int64    `gorm:"column:user_id" json:"-"`
	DishId    int64    `gorm:"column:dish_id" json:"-"`
	Guid      string    `gorm:"column:guid" json:"guid"`
	CreatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Image) TableName() string {
	return "images"
}
