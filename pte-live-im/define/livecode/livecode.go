package livecode

// 电商直播业务消息码 11001–11028
const (
	ProductExplainStart   = 11001
	ProductExplainCancel  = 11002
	Danmaku               = 11003
	GiftSend              = 11004
	ProductOnShelf        = 11005
	ProductOffShelf       = 11006
	MuteAll               = 11007
	UnmuteAll             = 11008
	MuteUser              = 11009
	UnmuteUser            = 11010
	DanmakuAuditOn        = 11011
	DanmakuAuditOff       = 11012
	LinkMicApply          = 11013
	LinkMicAgree          = 11014
	LinkMicReject         = 11015
	KickUser              = 11016
	BroadcastStart        = 11017
	BroadcastEnd          = 11018
	StreamStart           = 11019
	StreamInterrupt       = 11020
	StatusChange          = 11021
	OnlineCount           = 11022
	TotalCount            = 11023
	ConfigChange          = 11024
	LikeUpdate            = 11025
	RedpackEvent            = 11026
	DanmakuAuditPending     = 11027
	UserEnterWelcome        = 11028 // 进房欢迎（WS 连接后广播）
)

// LiveStatus 直播状态
const (
	StatusNotStarted = 0
	StatusLiving     = 1
	StatusEnded      = 2
	StatusAway       = 3
)

// GroupPrefix 电商直播分组前缀
const GroupPrefix = "live:"

func GroupName(roomId string) string {
	return GroupPrefix + roomId
}
