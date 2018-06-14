package models

import "time"

type Message struct {
	Id         int64    `gorm:"column:id" gorm:"primary_key" json:"id"`
	OwnerId    int64    `gorm:"column:owner_id" json:"ownerId"`
	ProviderId int64    `gorm:"column:provider_id" json:"providerId"`
	DishId     int64    `gorm:"column:dish_id" json:"dishId"`
	Name       string    `gorm:"column:name" json:"name"`
	Body       string    `gorm:"column:body" json:"body"`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt  time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Message) TableName() string {
	return "messages"
}
