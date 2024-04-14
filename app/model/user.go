package model

type GoogleSignIn struct {
	UserNo int `gorm:"primaryKey"`
	UserId string
	Token  string
}
