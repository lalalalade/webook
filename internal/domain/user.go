package domain

import "time"

// User 领域对象，是 DDD 中的 entity
type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	Ctime      time.Time
	WechatInfo WechatInfo
	Nickname   string
	Info       string
	Birthday   time.Time
}
