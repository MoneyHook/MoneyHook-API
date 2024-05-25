package user

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	UserExists(userId *string) *int
	UpdateToken(googleSignIn *model.GoogleSignIn)
	CreateUser(googleSignIn *model.GoogleSignIn) *model.GoogleSignIn
	CreateToken(googleSignIn *model.GoogleSignIn)
	ExtractUserNoFromToken(userToken *string) (*int, error)
	ExtractUserNoFromUserId(userId *string) (*int, error)
}
