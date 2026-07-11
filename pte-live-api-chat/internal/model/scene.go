package model

import "time"

const (
	SceneTypeShow  = "show"
	SceneTypeVoice = "voice"

	SceneRoomStatusPreparing = 1
	SceneRoomStatusLive      = 2
	SceneRoomStatusEnded     = 3
	SceneRoomStatusClosed    = 4

	SceneMemberStatusOnline  = 1
	SceneMemberStatusOffline = 2
	SceneMemberStatusKicked  = 3

	SceneMemberRoleOwner    = 1
	SceneMemberRoleAnchor   = 2
	SceneMemberRoleAdmin    = 3
	SceneMemberRoleAudience = 4

	SceneSeatStatusEmpty    = 1
	SceneSeatStatusOccupied = 2
	SceneSeatStatusLocked   = 3

	SceneMicStatusNormal = 1
	SceneMicStatusMuted  = 2

	SceneMicRequestStatusPending  = 1
	SceneMicRequestStatusAccepted = 2
	SceneMicRequestStatusRejected = 3
	SceneMicRequestStatusCanceled = 4
	SceneMicRequestStatusTimeout  = 5

	ScenePKStatusInviting = 1
	ScenePKStatusActive   = 2
	ScenePKStatusEnded    = 3
	ScenePKStatusCanceled = 4
	ScenePKStatusRejected = 5
	ScenePKStatusTimeout  = 6
)

type SceneRoom struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID     int       `gorm:"column:app_id;uniqueIndex:uniq_scene_room,priority:1;index:idx_scene_room_status,priority:1" json:"app_id"`
	SceneType string    `gorm:"column:scene_type;size:16;uniqueIndex:uniq_scene_room,priority:2;index:idx_scene_room_status,priority:2" json:"scene_type"`
	RoomID    string    `gorm:"column:room_id;size:96;uniqueIndex:uniq_scene_room,priority:3" json:"room_id"`
	Title     string    `gorm:"column:title;size:128" json:"title"`
	Cover     string    `gorm:"column:cover;size:512" json:"cover"`
	OwnerID   int64     `gorm:"column:owner_id;index:idx_scene_room_owner" json:"owner_id"`
	Status    int       `gorm:"column:status;default:1;index:idx_scene_room_status,priority:3" json:"status"`
	SeatCount int       `gorm:"column:seat_count" json:"seat_count"`
	Notice    string    `gorm:"column:notice;size:512" json:"notice"`
	Payload   string    `gorm:"column:payload;type:json" json:"payload"`
	StartedAt int64     `gorm:"column:started_at" json:"started_at"`
	EndedAt   int64     `gorm:"column:ended_at" json:"ended_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;index:idx_scene_room_status,priority:4" json:"updated_at"`
}

func (SceneRoom) TableName() string { return "scene_room" }

type SceneMember struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID      int       `gorm:"column:app_id;uniqueIndex:uniq_scene_member,priority:1;index:idx_scene_member_status,priority:1" json:"app_id"`
	SceneType  string    `gorm:"column:scene_type;size:16;uniqueIndex:uniq_scene_member,priority:2" json:"scene_type"`
	RoomID     string    `gorm:"column:room_id;size:96;uniqueIndex:uniq_scene_member,priority:3;index:idx_scene_member_room" json:"room_id"`
	UserID     int64     `gorm:"column:user_id;uniqueIndex:uniq_scene_member,priority:4;index:idx_scene_member_user" json:"user_id"`
	Role       int       `gorm:"column:role;default:4" json:"role"`
	Status     int       `gorm:"column:status;default:1;index:idx_scene_member_status,priority:2" json:"status"`
	Nickname   string    `gorm:"column:nickname;size:128" json:"nickname"`
	Avatar     string    `gorm:"column:avatar;size:512" json:"avatar"`
	MuteUntil  int64     `gorm:"column:mute_until" json:"mute_until"`
	JoinedAt   int64     `gorm:"column:joined_at" json:"joined_at"`
	LastSeenAt int64     `gorm:"column:last_seen_at" json:"last_seen_at"`
	LeftAt     int64     `gorm:"column:left_at" json:"left_at"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SceneMember) TableName() string { return "scene_member" }

type SceneSeat struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID     int       `gorm:"column:app_id;uniqueIndex:uniq_scene_seat,priority:1" json:"app_id"`
	SceneType string    `gorm:"column:scene_type;size:16;uniqueIndex:uniq_scene_seat,priority:2" json:"scene_type"`
	RoomID    string    `gorm:"column:room_id;size:96;uniqueIndex:uniq_scene_seat,priority:3;index:idx_scene_seat_room" json:"room_id"`
	SeatNo    int       `gorm:"column:seat_no;uniqueIndex:uniq_scene_seat,priority:4" json:"seat_no"`
	UserID    int64     `gorm:"column:user_id;index:idx_scene_seat_user" json:"user_id"`
	Status    int       `gorm:"column:status;default:1" json:"status"`
	MicStatus int       `gorm:"column:mic_status;default:1" json:"mic_status"`
	UpdatedBy int64     `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SceneSeat) TableName() string { return "scene_seat" }

type SceneMicRequest struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID      int       `gorm:"column:app_id;index:idx_scene_mic_room" json:"app_id"`
	SceneType  string    `gorm:"column:scene_type;size:16" json:"scene_type"`
	RoomID     string    `gorm:"column:room_id;size:96;index:idx_scene_mic_room" json:"room_id"`
	RequestID  string    `gorm:"column:request_id;size:96;uniqueIndex:uniq_scene_mic_request" json:"request_id"`
	Action     string    `gorm:"column:action;size:24" json:"action"`
	UserID     int64     `gorm:"column:user_id;index:idx_scene_mic_user" json:"user_id"`
	OperatorID int64     `gorm:"column:operator_id" json:"operator_id"`
	SeatNo     int       `gorm:"column:seat_no" json:"seat_no"`
	Status     int       `gorm:"column:status;default:1" json:"status"`
	Reason     string    `gorm:"column:reason;size:255" json:"reason"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (SceneMicRequest) TableName() string { return "scene_mic_request" }

type ScenePK struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID        int       `gorm:"column:app_id;index:idx_scene_pk_room" json:"app_id"`
	SceneType    string    `gorm:"column:scene_type;size:16" json:"scene_type"`
	RoomID       string    `gorm:"column:room_id;size:96;index:idx_scene_pk_room" json:"room_id"`
	PKID         string    `gorm:"column:pk_id;size:96;uniqueIndex:uniq_scene_pk" json:"pk_id"`
	TargetRoomID string    `gorm:"column:target_room_id;size:96" json:"target_room_id"`
	InviterID    int64     `gorm:"column:inviter_id" json:"inviter_id"`
	InviteeID    int64     `gorm:"column:invitee_id" json:"invitee_id"`
	Status       int       `gorm:"column:status;default:1" json:"status"`
	Score        string    `gorm:"column:score;type:json" json:"score"`
	StartedAt    int64     `gorm:"column:started_at" json:"started_at"`
	EndedAt      int64     `gorm:"column:ended_at" json:"ended_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ScenePK) TableName() string { return "scene_pk" }

type SceneEvent struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	AppID     int       `gorm:"column:app_id;index:idx_scene_event_room,priority:1" json:"app_id"`
	SceneType string    `gorm:"column:scene_type;size:16;index:idx_scene_event_room,priority:2" json:"scene_type"`
	RoomID    string    `gorm:"column:room_id;size:96;index:idx_scene_event_room,priority:3" json:"room_id"`
	GroupName string    `gorm:"column:group_name;size:128" json:"group_name"`
	EventType string    `gorm:"column:event_type;size:64" json:"event_type"`
	ActorID   int64     `gorm:"column:actor_id" json:"actor_id"`
	TargetID  int64     `gorm:"column:target_id" json:"target_id"`
	Code      int       `gorm:"column:code" json:"code"`
	Payload   string    `gorm:"column:payload;type:json" json:"payload"`
	CreatedAt time.Time `gorm:"column:created_at;index:idx_scene_event_room,priority:4" json:"created_at"`
}

func (SceneEvent) TableName() string { return "scene_event" }
