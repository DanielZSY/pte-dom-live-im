package retcode

const (
	SUCCESS = 0
	FAIL    = -1

	SYSTEM_ID_ERROR      = -1001 // 兼容旧码
	APP_ID_ERROR         = -1001
	ROOM_ID_ERROR        = -1002
	TOKEN_ERROR          = -1003
	KICKED_ERROR         = -1004
	MUTED_ERROR          = -1005
	ONLINE_MESSAGE_CODE  = 1001
	OFFLINE_MESSAGE_CODE = 1002
)
