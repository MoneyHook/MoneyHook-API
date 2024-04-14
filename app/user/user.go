package user

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	UserExists(userId *string) *int
	UpdateToken(googleSignIn *model.GoogleSignIn)
	CreateUser(googleSignIn *model.GoogleSignIn) *model.GoogleSignIn
	CreateToken(googleSignIn *model.GoogleSignIn)
}
