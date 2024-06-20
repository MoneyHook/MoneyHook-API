package fixed

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetFixedData(userId int) *[]model.GetFixed
	GetFixedDeletedData(userId int) *[]model.GetDeletedFixed
	AddFixed(*model.AddFixed) error
	EditFixed(*model.EditFixed) error
	DeleteFixed(*model.DeleteFixed) error
}
