package user

import "errors"

var ErrIdentityConflict = errors.New("firebase identity conflicts with a legacy user")

type Store interface {
	ResolveFirebaseUser(firebaseUID string, legacyUserID string) (string, error)
}
