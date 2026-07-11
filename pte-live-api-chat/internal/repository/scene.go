package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pte_live_api_chat/internal/model"
)

type SceneRepository struct {
	db *gorm.DB
}

func NewSceneRepository(db *gorm.DB) *SceneRepository {
	return &SceneRepository{db: db}
}

func (r *SceneRepository) Ready() bool {
	return r != nil && r.db != nil
}

func (r *SceneRepository) OpenRoom(ctx context.Context, req SceneRoomParams) (*model.SceneRoom, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	now := time.Now().Unix()
	var room model.SceneRoom
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seatCount := req.SeatCount
		if seatCount <= 0 {
			seatCount = defaultSeatCount(req.SceneType)
		}
		if seatCount > 32 {
			seatCount = 32
		}
		payload := normalizeJSON(req.Payload)
		room = model.SceneRoom{
			AppID:     req.AppID,
			SceneType: req.SceneType,
			RoomID:    req.RoomID,
			Title:     strings.TrimSpace(req.Title),
			Cover:     strings.TrimSpace(req.Cover),
			OwnerID:   req.OwnerID,
			Status:    model.SceneRoomStatusLive,
			SeatCount: seatCount,
			Notice:    strings.TrimSpace(req.Notice),
			Payload:   payload,
			StartedAt: now,
		}
		if room.Title == "" {
			room.Title = defaultRoomTitle(req.SceneType)
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "app_id"}, {Name: "scene_type"}, {Name: "room_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":      room.Title,
				"cover":      room.Cover,
				"owner_id":   room.OwnerID,
				"status":     room.Status,
				"seat_count": room.SeatCount,
				"notice":     room.Notice,
				"payload":    room.Payload,
				"started_at": gorm.Expr("IF(started_at = 0, ?, started_at)", now),
				"ended_at":   0,
			}),
		}).Create(&room).Error; err != nil {
			return err
		}
		if err := tx.Where("app_id = ? AND scene_type = ? AND room_id = ?", req.AppID, req.SceneType, req.RoomID).First(&room).Error; err != nil {
			return err
		}
		if err := ensureSeats(tx, req.AppID, req.SceneType, req.RoomID, room.SeatCount); err != nil {
			return err
		}
		if req.OwnerID > 0 {
			if err := upsertSceneMember(tx, SceneMemberParams{
				AppID: req.AppID, SceneType: req.SceneType, RoomID: req.RoomID, UserID: req.OwnerID,
				Role: model.SceneMemberRoleOwner, Nickname: req.Nickname, Avatar: req.Avatar,
			}); err != nil {
				return err
			}
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.room.opened", req.OwnerID, 0, sceneCode("scene.room.opened"), map[string]interface{}{
			"room_id":    req.RoomID,
			"scene_type": req.SceneType,
			"owner_id":   req.OwnerID,
			"title":      room.Title,
			"seat_count": room.SeatCount,
		})
	})
	return &room, err
}

func (r *SceneRepository) CloseRoom(ctx context.Context, req SceneRoomParams) (*model.SceneRoom, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	now := time.Now().Unix()
	var room model.SceneRoom
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("app_id = ? AND scene_type = ? AND room_id = ?", req.AppID, req.SceneType, req.RoomID).
			First(&room).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.SceneRoom{}).Where("id = ?", room.ID).
			Updates(map[string]interface{}{"status": model.SceneRoomStatusEnded, "ended_at": now}).Error; err != nil {
			return err
		}
		room.Status = model.SceneRoomStatusEnded
		room.EndedAt = now
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.room.closed", req.OperatorID, 0, sceneCode("scene.room.closed"), map[string]interface{}{
			"room_id":  req.RoomID,
			"ended_at": now,
		})
	})
	return &room, err
}

func (r *SceneRepository) ListRooms(ctx context.Context, req SceneListParams) ([]model.SceneRoom, int64, error) {
	if !r.Ready() {
		return nil, 0, ErrChatNotInitialized
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	q := r.db.WithContext(ctx).Model(&model.SceneRoom{}).
		Where("app_id = ? AND scene_type = ?", req.AppID, req.SceneType)
	if req.Status > 0 {
		q = q.Where("status = ?", req.Status)
	}
	if strings.TrimSpace(req.Keyword) != "" {
		q = q.Where("room_id = ? OR title LIKE ?", req.Keyword, "%"+req.Keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SceneRoom
	err := q.Order("updated_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *SceneRepository) ListEvents(ctx context.Context, req SceneEventListParams) ([]model.SceneEvent, int64, error) {
	if !r.Ready() {
		return nil, 0, ErrChatNotInitialized
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	q := r.db.WithContext(ctx).Model(&model.SceneEvent{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ?", req.AppID, req.SceneType, req.RoomID)
	if strings.TrimSpace(req.EventType) != "" {
		q = q.Where("event_type = ?", strings.TrimSpace(req.EventType))
	}
	if req.BeforeID > 0 {
		q = q.Where("id < ?", req.BeforeID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SceneEvent
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *SceneRepository) RoomDetail(ctx context.Context, appID int, sceneType, roomID string) (*model.SceneRoom, []model.SceneMember, []model.SceneSeat, *model.ScenePK, error) {
	if !r.Ready() {
		return nil, nil, nil, nil, ErrChatNotInitialized
	}
	var room model.SceneRoom
	if err := r.db.WithContext(ctx).Where("app_id = ? AND scene_type = ? AND room_id = ?", appID, sceneType, roomID).First(&room).Error; err != nil {
		return nil, nil, nil, nil, err
	}
	members, err := r.ListMembers(ctx, appID, sceneType, roomID, 1, 50)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	seats, err := r.ListSeats(ctx, appID, sceneType, roomID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var pk model.ScenePK
	err = r.db.WithContext(ctx).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND status IN ?", appID, sceneType, roomID, []int{model.ScenePKStatusInviting, model.ScenePKStatusActive}).
		Order("id DESC").First(&pk).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &room, members, seats, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return &room, members, seats, &pk, nil
}

func (r *SceneRepository) EnterRoom(ctx context.Context, req SceneMemberParams) (*model.SceneMember, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	limits, err := packageLimitsForApp(ctx, r.db, req.AppID)
	if err != nil {
		return nil, err
	}
	var member model.SceneMember
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		if req.UserID <= 0 {
			return errors.New("缺少 user_id")
		}
		if err := ensureSceneRoomCapacity(tx, req, limits); err != nil {
			return err
		}
		if req.Role <= 0 {
			req.Role = model.SceneMemberRoleAudience
		}
		if err := upsertSceneMember(tx, req); err != nil {
			return err
		}
		if err := tx.Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ?", req.AppID, req.SceneType, req.RoomID, req.UserID).First(&member).Error; err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.member.entered", req.UserID, 0, sceneCode("scene.member.entered"), map[string]interface{}{
			"user_id":  req.UserID,
			"nickname": req.Nickname,
			"role":     member.Role,
		})
	})
	return &member, err
}

func (r *SceneRepository) LeaveRoom(ctx context.Context, req SceneMemberParams) error {
	if !r.Ready() {
		return ErrChatNotInitialized
	}
	now := time.Now().Unix()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SceneMember{}).
			Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ?", req.AppID, req.SceneType, req.RoomID, req.UserID).
			Updates(map[string]interface{}{"status": model.SceneMemberStatusOffline, "left_at": now, "last_seen_at": now}).Error; err != nil {
			return err
		}
		if err := releaseUserSeats(tx, req.AppID, req.SceneType, req.RoomID, req.UserID, req.UserID); err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.member.left", req.UserID, 0, sceneCode("scene.member.left"), map[string]interface{}{
			"user_id": req.UserID,
			"left_at": now,
		})
	})
}

func (r *SceneRepository) ListMembers(ctx context.Context, appID int, sceneType, roomID string, page, pageSize int) ([]model.SceneMember, error) {
	page, pageSize = normalizePage(page, pageSize)
	var rows []model.SceneMember
	err := r.db.WithContext(ctx).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND status = ?", appID, sceneType, roomID, model.SceneMemberStatusOnline).
		Order("role ASC, last_seen_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, err
}

func (r *SceneRepository) ListSeats(ctx context.Context, appID int, sceneType, roomID string) ([]model.SceneSeat, error) {
	var rows []model.SceneSeat
	err := r.db.WithContext(ctx).Where("app_id = ? AND scene_type = ? AND room_id = ?", appID, sceneType, roomID).
		Order("seat_no ASC").Find(&rows).Error
	return rows, err
}

func (r *SceneRepository) SeatAction(ctx context.Context, req SceneSeatActionParams) ([]model.SceneSeat, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		return applySeatAction(tx, req)
	})
	if err != nil {
		return nil, err
	}
	return r.ListSeats(ctx, req.AppID, req.SceneType, req.RoomID)
}

func (r *SceneRepository) ModerationAction(ctx context.Context, req SceneModerationParams) (*model.SceneMember, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var member model.SceneMember
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		if req.TargetID <= 0 {
			return errors.New("缺少目标用户")
		}
		action := strings.ToLower(strings.TrimSpace(req.Action))
		now := time.Now().Unix()
		updates := map[string]interface{}{"last_seen_at": now}
		eventType := ""
		switch action {
		case "mute":
			if req.Duration <= 0 {
				req.Duration = 3600
			}
			updates["mute_until"] = now + req.Duration
			eventType = "scene.member.muted"
		case "unmute":
			updates["mute_until"] = 0
			eventType = "scene.member.unmuted"
		case "kick":
			updates["status"] = model.SceneMemberStatusKicked
			updates["left_at"] = now
			eventType = "scene.member.kicked"
			if err := releaseUserSeats(tx, req.AppID, req.SceneType, req.RoomID, req.TargetID, req.OperatorID); err != nil {
				return err
			}
		default:
			return errors.New("不支持的治理动作")
		}
		res := tx.Model(&model.SceneMember{}).
			Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ?", req.AppID, req.SceneType, req.RoomID, req.TargetID).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("成员不存在")
		}
		if err := tx.Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ?", req.AppID, req.SceneType, req.RoomID, req.TargetID).First(&member).Error; err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, eventType, req.OperatorID, req.TargetID, sceneCode(eventType), map[string]interface{}{
			"target_id":  req.TargetID,
			"duration":   req.Duration,
			"mute_until": member.MuteUntil,
			"reason":     strings.TrimSpace(req.Reason),
		})
	})
	return &member, err
}

func (r *SceneRepository) StartPK(ctx context.Context, req ScenePKParams) (*model.ScenePK, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	now := time.Now().Unix()
	var pk model.ScenePK
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		if req.PKID == "" {
			req.PKID = fmt.Sprintf("pk_%d_%s_%d", req.AppID, req.RoomID, time.Now().UnixNano())
		}
		pk = model.ScenePK{
			AppID: req.AppID, SceneType: req.SceneType, RoomID: req.RoomID, PKID: req.PKID,
			TargetRoomID: req.TargetRoomID, InviterID: req.InviterID, InviteeID: req.InviteeID,
			Status: model.ScenePKStatusActive, Score: normalizeJSON(req.Score), StartedAt: now,
		}
		if err := tx.Create(&pk).Error; err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.pk.started", req.InviterID, req.InviteeID, sceneCode("scene.pk.started"), map[string]interface{}{
			"pk_id":          pk.PKID,
			"target_room_id": req.TargetRoomID,
			"invitee_id":     req.InviteeID,
			"started_at":     now,
		})
	})
	return &pk, err
}

func (r *SceneRepository) InvitePK(ctx context.Context, req ScenePKParams) (*model.ScenePK, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var pk model.ScenePK
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		if strings.TrimSpace(req.PKID) == "" {
			req.PKID = fmt.Sprintf("pk_%d_%s_%d", req.AppID, req.RoomID, time.Now().UnixNano())
		}
		pk = model.ScenePK{
			AppID: req.AppID, SceneType: req.SceneType, RoomID: req.RoomID, PKID: req.PKID,
			TargetRoomID: req.TargetRoomID, InviterID: req.InviterID, InviteeID: req.InviteeID,
			Status: model.ScenePKStatusInviting, Score: normalizeJSON(req.Score),
		}
		if err := tx.Create(&pk).Error; err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.pk.invited", req.InviterID, req.InviteeID, sceneCode("scene.pk.invited"), map[string]interface{}{
			"pk_id":          pk.PKID,
			"target_room_id": req.TargetRoomID,
			"invitee_id":     req.InviteeID,
		})
	})
	return &pk, err
}

func (r *SceneRepository) PKAction(ctx context.Context, req ScenePKParams) (*model.ScenePK, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	now := time.Now().Unix()
	var pk model.ScenePK
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("app_id = ? AND scene_type = ? AND room_id = ?", req.AppID, req.SceneType, req.RoomID)
		if strings.TrimSpace(req.PKID) != "" {
			q = q.Where("pk_id = ?", strings.TrimSpace(req.PKID))
		} else {
			q = q.Where("status IN ?", []int{model.ScenePKStatusInviting, model.ScenePKStatusActive}).Order("id DESC")
		}
		if err := q.First(&pk).Error; err != nil {
			return err
		}
		action := strings.ToLower(strings.TrimSpace(req.Action))
		updates := map[string]interface{}{}
		eventType := ""
		switch action {
		case "accept":
			if pk.Status != model.ScenePKStatusInviting {
				return errors.New("PK 邀请状态不可接受")
			}
			updates["status"] = model.ScenePKStatusActive
			updates["started_at"] = now
			eventType = "scene.pk.accepted"
			pk.Status = model.ScenePKStatusActive
			pk.StartedAt = now
		case "reject":
			if pk.Status != model.ScenePKStatusInviting {
				return errors.New("PK 邀请状态不可拒绝")
			}
			updates["status"] = model.ScenePKStatusRejected
			updates["ended_at"] = now
			eventType = "scene.pk.rejected"
			pk.Status = model.ScenePKStatusRejected
			pk.EndedAt = now
		case "timeout":
			if pk.Status != model.ScenePKStatusInviting {
				return errors.New("PK 邀请状态不可超时")
			}
			updates["status"] = model.ScenePKStatusTimeout
			updates["ended_at"] = now
			eventType = "scene.pk.timeout"
			pk.Status = model.ScenePKStatusTimeout
			pk.EndedAt = now
		case "score":
			if pk.Status != model.ScenePKStatusActive {
				return errors.New("PK 未进行中")
			}
			updates["score"] = normalizeJSON(req.Score)
			eventType = "scene.pk.score.updated"
			pk.Score = normalizeJSON(req.Score)
		default:
			return errors.New("不支持的 PK 动作")
		}
		if len(updates) > 0 {
			if err := tx.Model(&model.ScenePK{}).Where("id = ?", pk.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, eventType, req.InviterID, req.InviteeID, sceneCode(eventType), map[string]interface{}{
			"pk_id":          pk.PKID,
			"target_room_id": pk.TargetRoomID,
			"score":          normalizeJSON(req.Score),
			"action":         action,
		})
	})
	return &pk, err
}

func (r *SceneRepository) EndPK(ctx context.Context, req ScenePKParams) (*model.ScenePK, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	now := time.Now().Unix()
	var pk model.ScenePK
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("app_id = ? AND scene_type = ? AND room_id = ?", req.AppID, req.SceneType, req.RoomID)
		if req.PKID != "" {
			q = q.Where("pk_id = ?", req.PKID)
		} else {
			q = q.Where("status IN ?", []int{model.ScenePKStatusInviting, model.ScenePKStatusActive}).Order("id DESC")
		}
		if err := q.First(&pk).Error; err != nil {
			return err
		}
		status := model.ScenePKStatusEnded
		if strings.EqualFold(req.Action, "cancel") {
			status = model.ScenePKStatusCanceled
		}
		eventType := "scene.pk.ended"
		if strings.EqualFold(req.Action, "settle") {
			eventType = "scene.pk.settled"
		}
		if status == model.ScenePKStatusCanceled {
			eventType = "scene.pk.canceled"
		}
		if err := tx.Model(&model.ScenePK{}).Where("id = ?", pk.ID).
			Updates(map[string]interface{}{"status": status, "score": normalizeJSON(req.Score), "ended_at": now}).Error; err != nil {
			return err
		}
		pk.Status = status
		pk.EndedAt = now
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, eventType, req.InviterID, req.InviteeID, sceneCode(eventType), map[string]interface{}{
			"pk_id":    pk.PKID,
			"score":    normalizeJSON(req.Score),
			"ended_at": now,
		})
	})
	return &pk, err
}

func (r *SceneRepository) SendEvent(ctx context.Context, req SceneEventParams) (*model.SceneEvent, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var event model.SceneEvent
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireSceneRoom(tx, req.AppID, req.SceneType, req.RoomID); err != nil {
			return err
		}
		return createSceneEventWithRow(tx, req.AppID, req.SceneType, req.RoomID, req.EventType, req.ActorID, req.TargetID, firstPositive(req.Code, sceneCode(req.EventType)), normalizeJSONToMap(req.Payload), &event)
	})
	return &event, err
}

func (r *SceneRepository) SceneEvent(ctx context.Context, appID int, eventID uint64) (*model.SceneEvent, error) {
	var row model.SceneEvent
	if err := r.db.WithContext(ctx).Where("app_id = ? AND id = ?", appID, eventID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *SceneRepository) ExpirePendingMicRequests(ctx context.Context, ttlSeconds int64, limit int) (int, error) {
	if !r.Ready() {
		return 0, ErrChatNotInitialized
	}
	if ttlSeconds <= 0 {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	cutoff := time.Now().Add(-time.Duration(ttlSeconds) * time.Second)
	var rows []model.SceneMicRequest
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", model.SceneMicRequestStatusPending, cutoff).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return 0, err
	}
	affected := 0
	for _, row := range rows {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&model.SceneMicRequest{}).
				Where("id = ? AND status = ?", row.ID, model.SceneMicRequestStatusPending).
				Update("status", model.SceneMicRequestStatusTimeout)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			affected++
			return createSceneEvent(tx, row.AppID, row.SceneType, row.RoomID, "scene.mic.timeout", row.OperatorID, row.UserID, sceneCode("scene.mic.timeout"), map[string]interface{}{
				"request_id": row.RequestID,
				"seat_no":    row.SeatNo,
				"reason":     "timeout",
			})
		})
		if err != nil {
			return affected, err
		}
	}
	return affected, nil
}

func (r *SceneRepository) ExpirePendingPKInvites(ctx context.Context, ttlSeconds int64, limit int) (int, error) {
	if !r.Ready() {
		return 0, ErrChatNotInitialized
	}
	if ttlSeconds <= 0 {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	cutoff := time.Now().Add(-time.Duration(ttlSeconds) * time.Second)
	var rows []model.ScenePK
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", model.ScenePKStatusInviting, cutoff).
		Order("id ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return 0, err
	}
	affected := 0
	for _, row := range rows {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&model.ScenePK{}).
				Where("id = ? AND status = ?", row.ID, model.ScenePKStatusInviting).
				Updates(map[string]interface{}{"status": model.ScenePKStatusTimeout, "ended_at": time.Now().Unix()})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}
			affected++
			return createSceneEvent(tx, row.AppID, row.SceneType, row.RoomID, "scene.pk.timeout", row.InviterID, row.InviteeID, sceneCode("scene.pk.timeout"), map[string]interface{}{
				"pk_id":          row.PKID,
				"target_room_id": row.TargetRoomID,
				"reason":         "timeout",
			})
		})
		if err != nil {
			return affected, err
		}
	}
	return affected, nil
}

func (r *ChatRepository) SceneEvent(ctx context.Context, appID int, eventID uint64) (*model.SceneEvent, error) {
	if !r.Ready() {
		return nil, ErrChatNotInitialized
	}
	var row model.SceneEvent
	if err := r.db.WithContext(ctx).Where("app_id = ? AND id = ?", appID, eventID).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

type SceneRoomParams struct {
	AppID      int
	SceneType  string
	RoomID     string
	Title      string
	Cover      string
	OwnerID    int64
	OperatorID int64
	SeatCount  int
	Notice     string
	Payload    string
	Nickname   string
	Avatar     string
}

type SceneMemberParams struct {
	AppID     int
	SceneType string
	RoomID    string
	UserID    int64
	Role      int
	Nickname  string
	Avatar    string
}

type SceneListParams struct {
	AppID     int
	SceneType string
	Status    int
	Keyword   string
	Page      int
	PageSize  int
}

type SceneEventListParams struct {
	AppID     int
	SceneType string
	RoomID    string
	EventType string
	BeforeID  uint64
	Page      int
	PageSize  int
}

type SceneSeatActionParams struct {
	AppID      int
	SceneType  string
	RoomID     string
	Action     string
	RequestID  string
	OperatorID int64
	UserID     int64
	TargetID   int64
	SeatNo     int
	Reason     string
}

type SceneModerationParams struct {
	AppID      int
	SceneType  string
	RoomID     string
	Action     string
	OperatorID int64
	TargetID   int64
	Duration   int64
	Reason     string
}

type ScenePKParams struct {
	AppID        int
	SceneType    string
	RoomID       string
	PKID         string
	TargetRoomID string
	InviterID    int64
	InviteeID    int64
	Action       string
	Score        string
}

type SceneEventParams struct {
	AppID     int
	SceneType string
	RoomID    string
	EventType string
	ActorID   int64
	TargetID  int64
	Code      int
	Payload   string
}

func ensureSeats(tx *gorm.DB, appID int, sceneType, roomID string, seatCount int) error {
	for i := 1; i <= seatCount; i++ {
		row := model.SceneSeat{AppID: appID, SceneType: sceneType, RoomID: roomID, SeatNo: i, Status: model.SceneSeatStatusEmpty, MicStatus: model.SceneMicStatusNormal}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func upsertSceneMember(tx *gorm.DB, req SceneMemberParams) error {
	now := time.Now().Unix()
	row := model.SceneMember{
		AppID: req.AppID, SceneType: req.SceneType, RoomID: req.RoomID, UserID: req.UserID,
		Role: req.Role, Status: model.SceneMemberStatusOnline, Nickname: strings.TrimSpace(req.Nickname),
		Avatar: strings.TrimSpace(req.Avatar), JoinedAt: now, LastSeenAt: now,
	}
	if row.Role <= 0 {
		row.Role = model.SceneMemberRoleAudience
	}
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "app_id"}, {Name: "scene_type"}, {Name: "room_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":       model.SceneMemberStatusOnline,
			"role":         gorm.Expr("LEAST(role, ?)", row.Role),
			"nickname":     row.Nickname,
			"avatar":       row.Avatar,
			"last_seen_at": now,
			"left_at":      0,
		}),
	}).Create(&row).Error
}

func requireSceneRoom(tx *gorm.DB, appID int, sceneType, roomID string) error {
	var count int64
	if err := tx.Model(&model.SceneRoom{}).Where("app_id = ? AND scene_type = ? AND room_id = ? AND status IN ?", appID, sceneType, roomID, []int{model.SceneRoomStatusPreparing, model.SceneRoomStatusLive}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("房间不存在或已结束")
	}
	return nil
}

func applySeatAction(tx *gorm.DB, req SceneSeatActionParams) error {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	actorID := firstPositive64(req.OperatorID, req.UserID)
	targetID := firstPositive64(req.TargetID, req.UserID)
	switch action {
	case "apply", "invite":
		requestID := fmt.Sprintf("%s_%d_%s_%d_%d", action, req.AppID, req.RoomID, targetID, time.Now().UnixNano())
		row := model.SceneMicRequest{
			AppID: req.AppID, SceneType: req.SceneType, RoomID: req.RoomID, RequestID: requestID,
			Action: action, UserID: targetID, OperatorID: actorID, SeatNo: req.SeatNo, Status: model.SceneMicRequestStatusPending,
			Reason: strings.TrimSpace(req.Reason),
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.mic."+action, actorID, targetID, sceneCode("scene.mic."+action), map[string]interface{}{
			"request_id": requestID,
			"seat_no":    req.SeatNo,
			"reason":     row.Reason,
		})
	case "take", "accept":
		if targetID <= 0 {
			return errors.New("缺少上麦用户")
		}
		if action == "accept" {
			if err := updateMicRequest(tx, req, targetID, model.SceneMicRequestStatusAccepted, "scene.mic.accepted"); err != nil {
				return err
			}
		}
		return occupySeat(tx, req, actorID, targetID)
	case "reject", "cancel", "timeout":
		if targetID <= 0 {
			return errors.New("缺少上麦用户")
		}
		status := model.SceneMicRequestStatusRejected
		eventType := "scene.mic.rejected"
		if action == "cancel" {
			status = model.SceneMicRequestStatusCanceled
			eventType = "scene.mic.canceled"
		}
		if action == "timeout" {
			status = model.SceneMicRequestStatusTimeout
			eventType = "scene.mic.timeout"
		}
		return updateMicRequest(tx, req, targetID, status, eventType)
	case "leave", "kick":
		if targetID <= 0 {
			return errors.New("缺少下麦用户")
		}
		if err := releaseUserSeats(tx, req.AppID, req.SceneType, req.RoomID, targetID, actorID); err != nil {
			return err
		}
		return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.seat."+action, actorID, targetID, sceneCode("scene.seat."+action), map[string]interface{}{
			"seat_no": req.SeatNo,
		})
	case "lock", "unlock", "mute", "unmute":
		return updateSeatFlag(tx, req, actorID, action)
	default:
		return errors.New("不支持的麦位动作")
	}
}

func updateMicRequest(tx *gorm.DB, req SceneSeatActionParams, targetID int64, status int, eventType string) error {
	q := tx.Model(&model.SceneMicRequest{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ? AND status = ?", req.AppID, req.SceneType, req.RoomID, targetID, model.SceneMicRequestStatusPending)
	if strings.TrimSpace(req.RequestID) != "" {
		q = q.Where("request_id = ?", strings.TrimSpace(req.RequestID))
	}
	res := q.Order("id DESC").Limit(1).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 && status != model.SceneMicRequestStatusAccepted {
		return errors.New("待处理上麦申请不存在")
	}
	actorID := firstPositive64(req.OperatorID, req.UserID)
	return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, eventType, actorID, targetID, sceneCode(eventType), map[string]interface{}{
		"request_id": strings.TrimSpace(req.RequestID),
		"seat_no":    req.SeatNo,
		"reason":     strings.TrimSpace(req.Reason),
	})
}

func occupySeat(tx *gorm.DB, req SceneSeatActionParams, actorID, targetID int64) error {
	seatNo := req.SeatNo
	if seatNo <= 0 {
		var seat model.SceneSeat
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("app_id = ? AND scene_type = ? AND room_id = ? AND status = ?", req.AppID, req.SceneType, req.RoomID, model.SceneSeatStatusEmpty).
			Order("seat_no ASC").First(&seat).Error; err != nil {
			return err
		}
		seatNo = seat.SeatNo
	}
	var seat model.SceneSeat
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND seat_no = ?", req.AppID, req.SceneType, req.RoomID, seatNo).
		First(&seat).Error; err != nil {
		return err
	}
	if seat.Status == model.SceneSeatStatusLocked {
		return errors.New("麦位已锁定")
	}
	if seat.Status == model.SceneSeatStatusOccupied && seat.UserID != targetID {
		return errors.New("麦位已被占用")
	}
	if err := releaseUserSeats(tx, req.AppID, req.SceneType, req.RoomID, targetID, actorID); err != nil {
		return err
	}
	if err := tx.Model(&model.SceneSeat{}).Where("id = ?", seat.ID).Updates(map[string]interface{}{
		"user_id": targetID, "status": model.SceneSeatStatusOccupied, "mic_status": model.SceneMicStatusNormal, "updated_by": actorID,
	}).Error; err != nil {
		return err
	}
	return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, "scene.seat.taken", actorID, targetID, sceneCode("scene.seat.taken"), map[string]interface{}{
		"seat_no": seatNo,
		"user_id": targetID,
	})
}

func releaseUserSeats(tx *gorm.DB, appID int, sceneType, roomID string, userID, operatorID int64) error {
	return tx.Model(&model.SceneSeat{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ? AND status = ?", appID, sceneType, roomID, userID, model.SceneSeatStatusOccupied).
		Updates(map[string]interface{}{"user_id": 0, "status": model.SceneSeatStatusEmpty, "mic_status": model.SceneMicStatusNormal, "updated_by": operatorID}).Error
}

func updateSeatFlag(tx *gorm.DB, req SceneSeatActionParams, actorID int64, action string) error {
	if req.SeatNo <= 0 {
		return errors.New("缺少 seat_no")
	}
	updates := map[string]interface{}{"updated_by": actorID}
	eventType := "scene.seat." + action
	switch action {
	case "lock":
		updates["status"] = model.SceneSeatStatusLocked
		updates["user_id"] = 0
	case "unlock":
		updates["status"] = model.SceneSeatStatusEmpty
		updates["user_id"] = 0
	case "mute":
		updates["mic_status"] = model.SceneMicStatusMuted
	case "unmute":
		updates["mic_status"] = model.SceneMicStatusNormal
	}
	res := tx.Model(&model.SceneSeat{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND seat_no = ?", req.AppID, req.SceneType, req.RoomID, req.SeatNo).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("麦位不存在")
	}
	return createSceneEvent(tx, req.AppID, req.SceneType, req.RoomID, eventType, actorID, req.TargetID, sceneCode(eventType), map[string]interface{}{
		"seat_no": req.SeatNo,
	})
}

func createSceneEvent(tx *gorm.DB, appID int, sceneType, roomID, eventType string, actorID, targetID int64, code int, payload map[string]interface{}) error {
	var event model.SceneEvent
	return createSceneEventWithRow(tx, appID, sceneType, roomID, eventType, actorID, targetID, code, payload, &event)
}

func createSceneEventWithRow(tx *gorm.DB, appID int, sceneType, roomID, eventType string, actorID, targetID int64, code int, payload map[string]interface{}, event *model.SceneEvent) error {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["scene_type"] = sceneType
	payload["room_id"] = roomID
	raw, _ := json.Marshal(payload)
	row := model.SceneEvent{
		AppID: appID, SceneType: sceneType, RoomID: roomID, GroupName: sceneGroupName(sceneType, roomID),
		EventType: eventType, ActorID: actorID, TargetID: targetID, Code: code, Payload: string(raw),
	}
	if err := tx.Create(&row).Error; err != nil {
		return err
	}
	*event = row
	return createOutbox(tx, appID, eventType, row.ID, map[string]interface{}{
		"scene_event_id": row.ID,
		"group_name":     row.GroupName,
	})
}

func sceneGroupName(sceneType, roomID string) string {
	return sceneType + ":" + roomID
}

func ensureSceneRoomCapacity(tx *gorm.DB, req SceneMemberParams, limits IMQuotaLimits) error {
	limit := int64(limits.MaxVoiceRoomOnline)
	if limit <= 0 {
		limit = DefaultMaxVoiceRoomOnline
	}
	var existing int64
	if err := tx.Model(&model.SceneMember{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND user_id = ? AND status = ?", req.AppID, req.SceneType, req.RoomID, req.UserID, model.SceneMemberStatusOnline).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	var online int64
	if err := tx.Model(&model.SceneMember{}).
		Where("app_id = ? AND scene_type = ? AND room_id = ? AND status = ?", req.AppID, req.SceneType, req.RoomID, model.SceneMemberStatusOnline).
		Count(&online).Error; err != nil {
		return err
	}
	if online >= limit {
		return quotaExceededError("房间在线人数", online, limit)
	}
	return nil
}

func firstPositive(values ...int) int {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func firstPositive64(values ...int64) int64 {
	for _, v := range values {
		if v > 0 {
			return v
		}
	}
	return 0
}

func defaultSeatCount(sceneType string) int {
	if sceneType == model.SceneTypeVoice {
		return 8
	}
	return 4
}

func defaultRoomTitle(sceneType string) string {
	if sceneType == model.SceneTypeVoice {
		return "语音房"
	}
	return "直播间"
}

func normalizeJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "{}"
	}
	if !json.Valid([]byte(raw)) {
		return "{}"
	}
	return raw
}

func normalizeJSONToMap(raw string) map[string]interface{} {
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(normalizeJSON(raw)), &out); err != nil || out == nil {
		return map[string]interface{}{}
	}
	return out
}

func sceneCode(eventType string) int {
	switch eventType {
	case "scene.room.opened":
		return 12001
	case "scene.room.closed":
		return 12002
	case "scene.member.entered":
		return 12003
	case "scene.member.left":
		return 12004
	case "scene.mic.apply":
		return 12011
	case "scene.mic.invite":
		return 12012
	case "scene.mic.accepted":
		return 12013
	case "scene.mic.rejected", "scene.mic.canceled", "scene.mic.timeout":
		return 12017
	case "scene.seat.taken":
		return 12013
	case "scene.seat.leave", "scene.seat.kick":
		return 12014
	case "scene.seat.lock", "scene.seat.unlock":
		return 12015
	case "scene.seat.mute", "scene.seat.unmute":
		return 12016
	case "scene.pk.invited":
		return 12020
	case "scene.pk.started", "scene.pk.accepted":
		return 12021
	case "scene.pk.ended", "scene.pk.canceled", "scene.pk.rejected", "scene.pk.timeout", "scene.pk.settled":
		return 12022
	case "scene.pk.score.updated":
		return 12023
	case "scene.gift.sent":
		return 12031
	case "scene.effect.played":
		return 12032
	case "scene.member.muted", "scene.member.unmuted":
		return 12041
	case "scene.member.kicked":
		return 12042
	default:
		return 12099
	}
}
