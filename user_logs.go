package golang_gorm

type UserLog struct {
	ID        int    `gorm:"primary_key;column:id;autoIncrement"`
	UserId    string `gorm:"column:user_id"`
	Action    string `gorm:"column:action"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (u *UserLog) TableName() string {
	return "user_logs"
}
