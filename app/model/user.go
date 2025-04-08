package model

type GoogleSignIn struct {
	UserNo string `gorm:"primaryKey"`
	UserId string
	Token  string
}
