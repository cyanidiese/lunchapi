package models

import "time"

type User struct {
	Id         int64    `gorm:"column:id" gorm:"primary_key" json:"id"`
	FirstName  string    `gorm:"column:first_name" json:"firstName"`
	LastName   string    `gorm:"column:last_name" json:"lastName"`
	Alias      string    `gorm:"column:alias" json:"alias"`
	Email      string    `gorm:"column:email" json:"email"`
	Token      string    `gorm:"column:token" json:"-"`
	IsProvider bool      `gorm:"column:is_provider" json:"isProvider"`
	IsShop     bool      `gorm:"column:is_shop" json:"isShop"`
	ProviderId int64    `gorm:"column:provider_id" json:"providerId"`
	OfficeId   int64    `gorm:"column:office_id" json:"-"`
	Office     Office    `gorm:"foreignkey:OfficeId;association_foreignkey:Id" json:"office"`
	RoleId     int64    `gorm:"column:role_id" json:"-"`
	Role       Role      `gorm:"foreignkey:RoleId;association_foreignkey:Id" json:"role"`
	Messages   []Message `gorm:"foreignkey:UserId;association_foreignkey:Id" json:"-"`
	Menus      []Menu    `gorm:"foreignkey:ProviderId;association_foreignkey:Id" json:"-"`
	Image      Image     `gorm:"foreignkey:UserId;association_foreignkey:Id" json:"image"`
	Timezone   string    `gorm:"column:timezone" json:"timezone"`
	CreatedAt  time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt  time.Time `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (User) TableName() string {
	return "users"
}
