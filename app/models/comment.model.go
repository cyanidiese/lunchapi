package models

import "time"

type Comment struct {
	Id        int64     `gorm:"column:id" gorm:"primary_key" json:"id"`
	OwnerId   int64     `gorm:"column:owner_id" json:"ownerId"`
	ItemId    int64     `gorm:"column:item_id" json:"itemId"`
	DishId    int64     `gorm:"column:dish_id" json:"dishId"`
	RepliedId int64     `gorm:"column:replied_id" json:"repliedId"`
	Body      string    `gorm:"column:body" json:"body"`
	Owner     User      `gorm:"foreignkey:OwnerId;association_foreignkey:Id" json:"owner"`
	CreatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Comment) TableName() string {
	return "comments"
}
