package service

import (
	"centralac/model"
	"centralac/serializer"
	"container/list"
	"sync"
)

// RoomWindStartService 请求送风的服务
type RoomWindStartService struct {
	RoomID string `form:"room_id" json:"room_id" binding:"required,min=3,max=4"`
}

var windSupplyLock sync.RWMutex
var windSupplySem uint = 3
var waitListLock sync.Mutex
var waitList = list.New()

// Start 请求送风函数
func (service *RoomWindStartService) Start() serializer.Response {
	var room model.Room
	if model.DB.Where("room_id = ?", service.RoomID).First(&room).RecordNotFound() {
		return serializer.ParamErr("房间号不存在", nil)
	}

	if room.WindSupply {
		return serializer.SystemErr("当前已在送风", nil)
	}

	centerStatusLock.RLock()
	if !centerPowerOn {
		centerStatusLock.RUnlock()
		return serializer.SystemErr("中央空调未开启", nil)
	}

	windSupplyLock.Lock()
	if windSupplySem > 0 {
		//开始送风
		windSupplySem--
		windSupplyLock.Unlock()
		centerStatusLock.RUnlock()
		return windSupply(&room)
	}

	waitListLock.Lock()
	waitList.PushBack(service.RoomID)
	waitListLock.Unlock()
	windSupplyLock.Unlock()
	centerStatusLock.RUnlock()
	resp := serializer.BuildRoomResponse(room)
	resp.Msg = "送风阻塞"
	return resp
}