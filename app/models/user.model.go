package models

import "time"

type User struct {
	Id              int64              `gorm:"column:id" gorm:"primary_key" json:"id"`
	FirstName       string             `gorm:"column:first_name" json:"firstName"`
	LastName        string             `gorm:"column:last_name" json:"lastName"`
	Alias           string             `gorm:"column:alias" json:"alias"`
	Email           string             `gorm:"column:email" json:"email"`
	Password        string             `gorm:"column:password" json:"-"`
	Token           string             `gorm:"column:token" json:"-"`
	IsProvider      bool               `gorm:"column:is_provider" json:"isProvider"`
	IsShop          bool               `gorm:"column:is_shop" json:"isShop"`
	ProviderId      int64              `gorm:"column:provider_id" json:"providerId"`
	OfficeId        int64              `gorm:"column:office_id" json:"officeId"`
	Office          Office             `gorm:"foreignkey:OfficeId;association_foreignkey:Id" json:"office"`
	RoleId          int64              `gorm:"column:role_id" json:"-"`
	Role            Role               `gorm:"foreignkey:RoleId;association_foreignkey:Id" json:"role"`
	Comments        []Comment          `gorm:"foreignkey:OwnerId;association_foreignkey:Id" json:"-"`
	Menus           []Menu             `gorm:"foreig4nkey:ProviderId;association_foreignkey:Id" json:"-"`
	Image           Image              `gorm:"foreignkey:UserId;association_foreignkey:Id" json:"image"`
	Timezone        string             `gorm:"column:timezone" json:"timezone"`
	Language        string             `gorm:"column:lang" sql:"DEFAULT:'en'" json:"lang"`
	IsDisabled      bool               `gorm:"column:is_disabled" json:"-"`
	MyNotifications []Notification     `gorm:"foreignkey:OwnerId;association_foreignkey:Id" json:"-"`
	Notifications   []UserNotification `gorm:"foreignkey:UserId;association_foreignkey:Id" json:"-"`
	CreatedAt       time.Time          `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt       time.Time          `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (User) TableName() string {
	return "users"
}
