package models

import "time"

type UserNotification struct {
	Id             int64        `gorm:"column:id" gorm:"primary_key" json:"id"`
	UserId         int64        `gorm:"column:user_id" json:"userId"`
	NotificationId int64        `gorm:"column:notification_id" json:"notificationId"`
	IsRead         bool         `gorm:"column:is_read" json:"isRead"`
	User           User         `gorm:"foreignkey:UserId;association_foreignkey:Id" json:"user"`
	Notification   Notification `gorm:"foreignkey:NotificationId;association_foreignkey:Id" json:"notification"`
	CreatedAt      time.Time    `sql:"DEFAULT:current_timestamp" json:"-"`
	UpdatedAt      time.Time    `sql:"DEFAULT:current_timestamp" json:"-"`
}

func (UserNotification) TableName() string {
	return "user_notifications"
}
