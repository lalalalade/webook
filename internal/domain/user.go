package domain

import "time"

// User 领域对象，是 DDD 中的 entity
type User struct {
	Email    string
	Password string
	Ctime    time.Time
}
