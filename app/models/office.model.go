package models

import "time"

type Office struct {
	Id         int64      `gorm:"column:id" gorm:"primary_key" json:"id"`
	TitleId    int64      `gorm:"column:title_id" json:"-"`
	Title      Translation `gorm:"foreignkey:TitleId;association_foreignkey:Id" json:"title"`
	Phone      string      `gorm:"column:phone" json:"phone"`
	Lat        float64     `gorm:"column:lat" json:"lat"`
	Lng        float64     `gorm:"column:lng" json:"lng"`
	Address    string      `gorm:"column:address" json:"address"`
	IsProvider bool        `gorm:"column:is_provider" json:"-"`
	CreatedAt  time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt  time.Time   `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Office) TableName() string {
	return "offices"
}
