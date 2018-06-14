package models

import "time"

type Dish struct {
	Id            int64      `gorm:"column:id" gorm:"primary_key" json:"id"`
	CategoryId    int64      `gorm:"column:category_id" json:"categoryId"`
	ProviderId    int64      `gorm:"column:provider_id" json:"providerId"`
	NameId        int64      `gorm:"column:name_id" json:"-"`
	Name          Translation `gorm:"foreignkey:NameId;association_foreignkey:Id" json:"name"`
	DescriptionId int64      `gorm:"column:description_id" json:"-"`
	Description   Translation `gorm:"foreignkey:DescriptionId;association_foreignkey:Id" json:"description"`
	Weight        float64     `gorm:"column:weight" json:"weight"`
	Calories      float64     `gorm:"column:calories" json:"calories"`
	Price         float64     `gorm:"column:price" json:"price"`
	Category      Category    `gorm:"foreignkey:CategoryId;association_foreignkey:Id" json:"-"`
	Provider      User        `gorm:"foreignkey:ProviderId;association_foreignkey:Id" json:"-"`
	Images        []Image     `gorm:"foreignkey:DishId;association_foreignkey:Id" json:"images"`
	CreatedAt     time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt     time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Dish) TableName() string {
	return "dishes"
}
