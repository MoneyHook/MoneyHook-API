package transaction

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetTimelineData(userId int, month string) *[]model.Timeline
}
