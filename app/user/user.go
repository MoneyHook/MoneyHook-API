package user

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	UserExists(userId *string) *string
	UpdateToken(googleSignIn *model.GoogleSignIn)
	CreateUser(googleSignIn *model.GoogleSignIn) *model.GoogleSignIn
	CreateToken(googleSignIn *model.GoogleSignIn)
	ExtractUserNoFromToken(userToken *string) (*string, error)
	ExtractUserNoFromUserId(userId *string) (*string, error)
}
