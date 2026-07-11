package service

import (
	"context"
	"errors"
	"strings"

	"pte_live_api_chat/internal/model"
	"pte_live_api_chat/internal/repository"
)

type SceneService struct {
	repo *repository.SceneRepository
}

func NewSceneService(repo *repository.SceneRepository) *SceneService {
	return &SceneService{repo: repo}
}

func (s *SceneService) Ready() bool {
	return s != nil && s.repo != nil && s.repo.Ready()
}

type SceneRoomRequest struct {
	AppID      int    `json:"app_id"`
	SceneType  string `json:"scene_type"`
	RoomID     string `json:"room_id"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	OwnerID    int64  `json:"owner_id"`
	AnchorID   int64  `json:"anchor_id"`
	OperatorID int64  `json:"operator_id"`
	SeatCount  int    `json:"seat_count"`
	Notice     string `json:"notice"`
	Payload    string `json:"payload"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
}

type SceneListRequest struct {
	AppID     int    `json:"app_id"`
	SceneType string `json:"scene_type"`
	Status    int    `json:"status"`
	Keyword   string `json:"keyword"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type SceneMemberRequest struct {
	AppID     int    `json:"app_id"`
	SceneType string `json:"scene_type"`
	RoomID    string `json:"room_id"`
	UserID    int64  `json:"user_id"`
	Role      int    `json:"role"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type SceneSeatActionRequest struct {
	AppID      int    `json:"app_id"`
	SceneType  string `json:"scene_type"`
	RoomID     string `json:"room_id"`
	Action     string `json:"action"`
	RequestID  string `json:"request_id"`
	OperatorID int64  `json:"operator_id"`
	UserID     int64  `json:"user_id"`
	TargetID   int64  `json:"target_id"`
	SeatNo     int    `json:"seat_no"`
	Reason     string `json:"reason"`
}

type SceneModerationRequest struct {
	AppID      int    `json:"app_id"`
	SceneType  string `json:"scene_type"`
	RoomID     string `json:"room_id"`
	Action     string `json:"action"`
	OperatorID int64  `json:"operator_id"`
	UserID     int64  `json:"user_id"`
	TargetID   int64  `json:"target_id"`
	Duration   int64  `json:"duration"`
	Reason     string `json:"reason"`
}

type ScenePKRequest struct {
	AppID        int    `json:"app_id"`
	SceneType    string `json:"scene_type"`
	RoomID       string `json:"room_id"`
	PKID         string `json:"pk_id"`
	TargetRoomID string `json:"target_room_id"`
	InviterID    int64  `json:"inviter_id"`
	InviteeID    int64  `json:"invitee_id"`
	Action       string `json:"action"`
	Score        string `json:"score"`
}

type SceneEventRequest struct {
	AppID     int    `json:"app_id"`
	SceneType string `json:"scene_type"`
	RoomID    string `json:"room_id"`
	EventType string `json:"event_type"`
	ActorID   int64  `json:"actor_id"`
	UserID    int64  `json:"user_id"`
	TargetID  int64  `json:"target_id"`
	Code      int    `json:"code"`
	Payload   string `json:"payload"`
}

type SceneEventListRequest struct {
	AppID     int    `json:"app_id"`
	SceneType string `json:"scene_type"`
	RoomID    string `json:"room_id"`
	EventType string `json:"event_type"`
	BeforeID  uint64 `json:"before_id"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type SceneRoomView struct {
	ID        uint64            `json:"id"`
	AppID     int               `json:"app_id"`
	SceneType string            `json:"scene_type"`
	RoomID    string            `json:"room_id"`
	GroupName string            `json:"group_name"`
	Title     string            `json:"title"`
	Cover     string            `json:"cover"`
	OwnerID   int64             `json:"owner_id"`
	Status    int               `json:"status"`
	SeatCount int               `json:"seat_count"`
	Notice    string            `json:"notice"`
	Payload   string            `json:"payload"`
	StartedAt int64             `json:"started_at"`
	EndedAt   int64             `json:"ended_at"`
	Members   []SceneMemberView `json:"members,omitempty"`
	Seats     []SceneSeatView   `json:"seats,omitempty"`
	PK        *ScenePKView      `json:"pk,omitempty"`
}

type SceneMemberView struct {
	UserID     int64  `json:"user_id"`
	Role       int    `json:"role"`
	Status     int    `json:"status"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	MuteUntil  int64  `json:"mute_until"`
	JoinedAt   int64  `json:"joined_at"`
	LastSeenAt int64  `json:"last_seen_at"`
}

type SceneSeatView struct {
	SeatNo    int   `json:"seat_no"`
	UserID    int64 `json:"user_id"`
	Status    int   `json:"status"`
	MicStatus int   `json:"mic_status"`
	UpdatedBy int64 `json:"updated_by"`
}

type ScenePKView struct {
	PKID         string `json:"pk_id"`
	TargetRoomID string `json:"target_room_id"`
	InviterID    int64  `json:"inviter_id"`
	InviteeID    int64  `json:"invitee_id"`
	Status       int    `json:"status"`
	Score        string `json:"score"`
	StartedAt    int64  `json:"started_at"`
	EndedAt      int64  `json:"ended_at"`
}

type SceneEventView struct {
	EventID   uint64 `json:"event_id"`
	GroupName string `json:"group_name"`
	EventType string `json:"event_type"`
	ActorID   int64  `json:"actor_id"`
	TargetID  int64  `json:"target_id"`
	Code      int    `json:"code"`
	Payload   string `json:"payload"`
}

type SceneRoomListResult struct {
	Total int64           `json:"total"`
	List  []SceneRoomView `json:"list"`
}

type SceneMemberListResult struct {
	List []SceneMemberView `json:"list"`
}

type SceneSeatListResult struct {
	List []SceneSeatView `json:"list"`
}

type SceneEventListResult struct {
	Total int64            `json:"total"`
	List  []SceneEventView `json:"list"`
}

func (s *SceneService) OpenRoom(ctx context.Context, appID int, req SceneRoomRequest) (*SceneRoomView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	ownerID := firstPositive64(req.OwnerID, req.AnchorID, req.OperatorID)
	if ownerID <= 0 {
		return nil, errors.New("缺少主播/房主")
	}
	room, err := s.repo.OpenRoom(ctx, repository.SceneRoomParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, Title: req.Title, Cover: req.Cover,
		OwnerID: ownerID, OperatorID: firstPositive64(req.OperatorID, ownerID), SeatCount: req.SeatCount,
		Notice: req.Notice, Payload: req.Payload, Nickname: req.Nickname, Avatar: req.Avatar,
	})
	if err != nil {
		return nil, err
	}
	return s.Detail(ctx, appID, SceneRoomRequest{SceneType: sceneType, RoomID: room.RoomID})
}

func (s *SceneService) CloseRoom(ctx context.Context, appID int, req SceneRoomRequest) (*SceneRoomView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	room, err := s.repo.CloseRoom(ctx, repository.SceneRoomParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, OperatorID: firstPositive64(req.OperatorID, req.OwnerID, req.AnchorID),
	})
	if err != nil {
		return nil, err
	}
	view := sceneRoomView(*room)
	return &view, nil
}

func (s *SceneService) ListRooms(ctx context.Context, appID int, req SceneListRequest) (*SceneRoomListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, _, err := normalizeScene(req.SceneType, "placeholder")
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repo.ListRooms(ctx, repository.SceneListParams{
		AppID: appID, SceneType: sceneType, Status: req.Status, Keyword: strings.TrimSpace(req.Keyword),
		Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SceneRoomView, 0, len(rows))
	for _, row := range rows {
		out = append(out, sceneRoomView(row))
	}
	return &SceneRoomListResult{Total: total, List: out}, nil
}

func (s *SceneService) Detail(ctx context.Context, appID int, req SceneRoomRequest) (*SceneRoomView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	room, members, seats, pk, err := s.repo.RoomDetail(ctx, appID, sceneType, roomID)
	if err != nil {
		return nil, err
	}
	view := sceneRoomView(*room)
	view.Members = sceneMemberViews(members)
	view.Seats = sceneSeatViews(seats)
	if pk != nil {
		pkView := scenePKView(*pk)
		view.PK = &pkView
	}
	return &view, nil
}

func (s *SceneService) EnterRoom(ctx context.Context, appID int, req SceneMemberRequest) (*SceneMemberView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	if req.UserID <= 0 {
		return nil, errors.New("缺少 user_id")
	}
	row, err := s.repo.EnterRoom(ctx, repository.SceneMemberParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, UserID: req.UserID, Role: req.Role,
		Nickname: req.Nickname, Avatar: req.Avatar,
	})
	if err != nil {
		return nil, err
	}
	view := sceneMemberView(*row)
	return &view, nil
}

func (s *SceneService) LeaveRoom(ctx context.Context, appID int, req SceneMemberRequest) (*OKResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	if req.UserID <= 0 {
		return nil, errors.New("缺少 user_id")
	}
	if err := s.repo.LeaveRoom(ctx, repository.SceneMemberParams{AppID: appID, SceneType: sceneType, RoomID: roomID, UserID: req.UserID}); err != nil {
		return nil, err
	}
	return &OKResult{OK: true}, nil
}

func (s *SceneService) ListMembers(ctx context.Context, appID int, req SceneMemberRequest) (*SceneMemberListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListMembers(ctx, appID, sceneType, roomID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &SceneMemberListResult{List: sceneMemberViews(rows)}, nil
}

func (s *SceneService) SeatAction(ctx context.Context, appID int, req SceneSeatActionRequest) (*SceneSeatListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	operatorID := firstPositive64(req.OperatorID, req.UserID)
	if operatorID <= 0 {
		return nil, errors.New("缺少 operator_id")
	}
	rows, err := s.repo.SeatAction(ctx, repository.SceneSeatActionParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, Action: req.Action, RequestID: req.RequestID, OperatorID: operatorID,
		UserID: req.UserID, TargetID: req.TargetID, SeatNo: req.SeatNo, Reason: req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return &SceneSeatListResult{List: sceneSeatViews(rows)}, nil
}

func (s *SceneService) ModerationAction(ctx context.Context, appID int, req SceneModerationRequest) (*SceneMemberView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	operatorID := firstPositive64(req.OperatorID, req.UserID)
	if operatorID <= 0 || req.TargetID <= 0 {
		return nil, errors.New("治理参数错误")
	}
	row, err := s.repo.ModerationAction(ctx, repository.SceneModerationParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, Action: req.Action,
		OperatorID: operatorID, TargetID: req.TargetID, Duration: req.Duration, Reason: req.Reason,
	})
	if err != nil {
		return nil, err
	}
	view := sceneMemberView(*row)
	return &view, nil
}

func (s *SceneService) StartPK(ctx context.Context, appID int, req ScenePKRequest) (*ScenePKView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	if req.TargetRoomID == "" || req.InviterID <= 0 {
		return nil, errors.New("PK 参数错误")
	}
	row, err := s.repo.StartPK(ctx, repository.ScenePKParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, PKID: req.PKID, TargetRoomID: req.TargetRoomID,
		InviterID: req.InviterID, InviteeID: req.InviteeID, Score: req.Score,
	})
	if err != nil {
		return nil, err
	}
	view := scenePKView(*row)
	return &view, nil
}

func (s *SceneService) InvitePK(ctx context.Context, appID int, req ScenePKRequest) (*ScenePKView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.TargetRoomID) == "" || req.InviterID <= 0 {
		return nil, errors.New("PK 邀请参数错误")
	}
	row, err := s.repo.InvitePK(ctx, repository.ScenePKParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, PKID: req.PKID,
		TargetRoomID: req.TargetRoomID, InviterID: req.InviterID, InviteeID: req.InviteeID, Score: req.Score,
	})
	if err != nil {
		return nil, err
	}
	view := scenePKView(*row)
	return &view, nil
}

func (s *SceneService) PKAction(ctx context.Context, appID int, req ScenePKRequest) (*ScenePKView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Action) == "" {
		return nil, errors.New("缺少 PK action")
	}
	row, err := s.repo.PKAction(ctx, repository.ScenePKParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, PKID: req.PKID,
		InviterID: req.InviterID, InviteeID: req.InviteeID, Action: req.Action, Score: req.Score,
	})
	if err != nil {
		return nil, err
	}
	view := scenePKView(*row)
	return &view, nil
}

func (s *SceneService) EndPK(ctx context.Context, appID int, req ScenePKRequest) (*ScenePKView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	row, err := s.repo.EndPK(ctx, repository.ScenePKParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, PKID: req.PKID, InviterID: req.InviterID,
		InviteeID: req.InviteeID, Action: req.Action, Score: req.Score,
	})
	if err != nil {
		return nil, err
	}
	view := scenePKView(*row)
	return &view, nil
}

func (s *SceneService) SendEvent(ctx context.Context, appID int, req SceneEventRequest) (*SceneEventView, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	eventType := strings.TrimSpace(req.EventType)
	if eventType == "" {
		eventType = "scene.message.sent"
	}
	actorID := firstPositive64(req.ActorID, req.UserID)
	row, err := s.repo.SendEvent(ctx, repository.SceneEventParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, EventType: eventType,
		ActorID: actorID, TargetID: req.TargetID, Code: req.Code, Payload: req.Payload,
	})
	if err != nil {
		return nil, err
	}
	view := sceneEventView(*row)
	return &view, nil
}

func (s *SceneService) ListEvents(ctx context.Context, appID int, req SceneEventListRequest) (*SceneEventListResult, error) {
	if !s.Ready() {
		return nil, repository.ErrChatNotInitialized
	}
	appID = normalizeAppID(firstPositive(appID, req.AppID))
	sceneType, roomID, err := normalizeScene(req.SceneType, req.RoomID)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.repo.ListEvents(ctx, repository.SceneEventListParams{
		AppID: appID, SceneType: sceneType, RoomID: roomID, EventType: req.EventType,
		BeforeID: req.BeforeID, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	out := make([]SceneEventView, 0, len(rows))
	for _, row := range rows {
		out = append(out, sceneEventView(row))
	}
	return &SceneEventListResult{Total: total, List: out}, nil
}

func normalizeScene(sceneType, roomID string) (string, string, error) {
	sceneType = strings.ToLower(strings.TrimSpace(sceneType))
	if sceneType != model.SceneTypeShow && sceneType != model.SceneTypeVoice {
		return "", "", errors.New("scene_type 仅支持 show/voice")
	}
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return "", "", errors.New("缺少 room_id")
	}
	return sceneType, roomID, nil
}

func sceneRoomView(row model.SceneRoom) SceneRoomView {
	return SceneRoomView{
		ID: row.ID, AppID: row.AppID, SceneType: row.SceneType, RoomID: row.RoomID, GroupName: row.SceneType + ":" + row.RoomID,
		Title: row.Title, Cover: row.Cover, OwnerID: row.OwnerID, Status: row.Status, SeatCount: row.SeatCount,
		Notice: row.Notice, Payload: row.Payload, StartedAt: row.StartedAt, EndedAt: row.EndedAt,
	}
}

func sceneMemberViews(rows []model.SceneMember) []SceneMemberView {
	out := make([]SceneMemberView, 0, len(rows))
	for _, row := range rows {
		out = append(out, sceneMemberView(row))
	}
	return out
}

func sceneMemberView(row model.SceneMember) SceneMemberView {
	return SceneMemberView{
		UserID: row.UserID, Role: row.Role, Status: row.Status, Nickname: row.Nickname, Avatar: row.Avatar,
		MuteUntil: row.MuteUntil, JoinedAt: row.JoinedAt, LastSeenAt: row.LastSeenAt,
	}
}

func sceneSeatViews(rows []model.SceneSeat) []SceneSeatView {
	out := make([]SceneSeatView, 0, len(rows))
	for _, row := range rows {
		out = append(out, SceneSeatView{SeatNo: row.SeatNo, UserID: row.UserID, Status: row.Status, MicStatus: row.MicStatus, UpdatedBy: row.UpdatedBy})
	}
	return out
}

func scenePKView(row model.ScenePK) ScenePKView {
	return ScenePKView{
		PKID: row.PKID, TargetRoomID: row.TargetRoomID, InviterID: row.InviterID, InviteeID: row.InviteeID,
		Status: row.Status, Score: row.Score, StartedAt: row.StartedAt, EndedAt: row.EndedAt,
	}
}

func sceneEventView(row model.SceneEvent) SceneEventView {
	return SceneEventView{
		EventID: row.ID, GroupName: row.GroupName, EventType: row.EventType, ActorID: row.ActorID,
		TargetID: row.TargetID, Code: row.Code, Payload: row.Payload,
	}
}
