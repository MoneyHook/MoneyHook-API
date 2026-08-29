package store_mysql

import (
	userdomain "MoneyHook/MoneyHook-API/user"
	"errors"
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserStore struct {
	db *gorm.DB
}

type userIdentity struct {
	UserNo string `gorm:"column:user_no"`
	UserID string `gorm:"column:user_id"`
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}

func (us *UserStore) ResolveFirebaseUser(firebaseUID string, legacyUserID string) (string, error) {
	var lastErr error
	for range 5 {
		userNo, err := us.resolveFirebaseUser(firebaseUID, legacyUserID)
		if err == nil {
			return userNo, nil
		}
		lastErr = err
		var mysqlErr *mysqldriver.MySQLError
		if !errors.As(err, &mysqlErr) || (mysqlErr.Number != 1213 && mysqlErr.Number != 1205) {
			return "", err
		}
	}
	return "", fmt.Errorf("resolve Firebase user after lock retries: %w", lastErr)
}

func (us *UserStore) resolveFirebaseUser(firebaseUID string, legacyUserID string) (string, error) {
	var userNo string
	err := us.db.Transaction(func(tx *gorm.DB) error {
		current, currentFound, err := findUserIdentity(tx, firebaseUID)
		if err != nil {
			return err
		}
		legacy, legacyFound, err := findUserIdentity(tx, legacyUserID)
		if err != nil {
			return err
		}
		if currentFound && legacyFound && current.UserNo != legacy.UserNo {
			return userdomain.ErrIdentityConflict
		}
		if currentFound {
			userNo = current.UserNo
			return nil
		}
		if legacyFound {
			result := tx.Table("users").
				Where("user_no = ? AND user_id = ?", legacy.UserNo, legacyUserID).
				Update("user_id", firebaseUID)
			if result.Error != nil {
				return fmt.Errorf("migrate legacy firebase identity: %w", result.Error)
			}
			if result.RowsAffected != 1 {
				return userdomain.ErrIdentityConflict
			}
			userNo = legacy.UserNo
			return nil
		}

		if err := tx.Table("users").Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"user_id"}),
		}).
			Create(map[string]any{"user_id": firebaseUID}).Error; err != nil {
			return fmt.Errorf("create firebase user: %w", err)
		}
		created, found, err := findUserIdentity(tx, firebaseUID)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("created firebase user could not be resolved")
		}
		userNo = created.UserNo
		return nil
	})
	return userNo, err
}

func findUserIdentity(db *gorm.DB, userID string) (userIdentity, bool, error) {
	var identity userIdentity
	err := db.Table("users").Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("user_no", "user_id").Where("user_id = ?", userID).Take(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return userIdentity{}, false, nil
	}
	if err != nil {
		return userIdentity{}, false, err
	}
	return identity, true, nil
}
