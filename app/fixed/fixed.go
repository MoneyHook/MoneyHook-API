package fixed

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetFixedData(userId string) *[]model.GetFixed
	GetFixedDeletedData(userId string) *[]model.GetDeletedFixed
	AddFixed(*model.AddFixed) error
	EditFixed(*model.EditFixed) error
	DeleteFixed(*model.DeleteFixed) error
}
