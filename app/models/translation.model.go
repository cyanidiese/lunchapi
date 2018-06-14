package models

import "time"

type Translation struct {
	Id        int64    `gorm:"column:id" gorm:"primary_key" json:"-"`
	En        string    `gorm:"column:en" sql:"type:text" json:"en"`
	Ua        string    `gorm:"column:ua" sql:"type:text" json:"ua"`
	Ru        string    `gorm:"column:ru" sql:"type:text" json:"ru"`
	CreatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Translation) TableName() string {
	return "translations"
}
