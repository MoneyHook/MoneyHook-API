package transaction

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetFixedData(userId int) *[]model.GetFixed
}
