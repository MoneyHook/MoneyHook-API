package fixed

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetFixedData(userId int) *[]model.GetFixed
	GetFixedDeletedData(userId int) *[]model.GetDeletedFixed
	AddFixed(*model.AddFixed)
	EditFixed(*model.EditFixed)
	DeleteFixed(*model.DeleteFixed)
}
