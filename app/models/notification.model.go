package models

import "time"

type Notification struct {
	Id         int64              `gorm:"column:id" gorm:"primary_key" json:"id"`
	OwnerId    int64              `gorm:"column:owner_id" json:"ownerId"`
	Body       string             `gorm:"column:body" json:"body"`
	IsApproved bool               `gorm:"column:is_approved" json:"isApproved"`
	UserNotes  []UserNotification `gorm:"foreignkey:NotificationId;association_foreignkey:Id" json:"userNotes"`
	CreatedAt  time.Time          `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt  time.Time          `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
