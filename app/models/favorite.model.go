package models

type Favorite struct {
	Id             int64    `gorm:"column:id" gorm:"primary_key" json:"id"`
	UserId         int64    `gorm:"column:user_id" json:"user_id"`
	DishId         int64    `gorm:"column:dish_id" json:"dishId"`
}

func (Favorite) TableName() string {
	return "favorites"
}
