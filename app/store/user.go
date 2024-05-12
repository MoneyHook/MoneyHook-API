package store

import (
	"MoneyHook/MoneyHook-API/model"
	"fmt"

	"gorm.io/gorm"
)

type UserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (us *UserStore) UserExists(userId *string) *int {
	var result int

	us.db.Table("users").
		Select("user_no").
		Where("user_id = ?", userId).
		Limit(1).
		Scan(&result)

	return &result
}

func (us *UserStore) UpdateToken(googleSignIn *model.GoogleSignIn) {
	subquery := us.db.Table("users").Select("user_no").Where("user_id = ?", googleSignIn.UserId)

	us.db.Table("user_token").
		Where("user_no = (?)", subquery).
		Update("token", googleSignIn.Token)
}

func (us *UserStore) CreateUser(googleSignIn *model.GoogleSignIn) *model.GoogleSignIn {
	us.db.Table("users").
		Model(&googleSignIn).
		Create(map[string]interface{}{
			"user_id": googleSignIn.UserId,
		})

	result := us.db.Table("users").
		Select("user_no").
		Where("user_id = ?", googleSignIn.UserId).
		Scan(&googleSignIn.UserNo)

	if result == nil {
		// エラー処理
		fmt.Println("error")
	}

	return googleSignIn
}

func (us *UserStore) CreateToken(googleSignIn *model.GoogleSignIn) {
	us.db.Table("user_token").Create(map[string]interface{}{
		"user_no": googleSignIn.UserNo,
		"token":   googleSignIn.Token,
	})
}

func (us *UserStore) ExtractUserNoFromToken(userToken *string) (*int, error) {
	model := model.GoogleSignIn{}
	result := us.db.Table("user_token").
		Select("user_no").
		Where("token = ?", userToken).
		Scan(&model)

	if result.Error == nil || model.UserNo == 0 {
		return &model.UserNo, gorm.ErrRecordNotFound
	}
	return &model.UserNo, nil
}

func (us *UserStore) ExtractUserNoFromUserId(userId *string) (*int, error) {
	model := model.GoogleSignIn{}
	result := us.db.Table("users").
		Select("user_no").
		Where("user_id = ?", "userId").
		Scan(&model)

	if result.Error == nil || model.UserNo == 0 {
		return &model.UserNo, gorm.ErrRecordNotFound
	}
	return &model.UserNo, nil
}
