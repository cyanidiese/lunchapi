package models

import "time"

type Role struct {
	Id        int64    `gorm:"column:id" gorm:"primary_key" json:"-"`
	Name      string    `gorm:"column:name" json:"name"`
	Title     string    `gorm:"column:title" json:"title"`
	CreatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}
