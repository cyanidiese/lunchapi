package models

import "time"

type Category struct {
	Id        int64      `gorm:"column:id" gorm:"primary_key" json:"id"`
	TitleId   int64      `gorm:"column:title_id" json:"-"`
	Title     Translation `gorm:"foreignkey:TitleId;association_foreignkey:Id" json:"title"`
	CreatedAt time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Category) TableName() string {
	return "categories"
}
