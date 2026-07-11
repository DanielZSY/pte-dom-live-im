package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"pte_live_api_chat/internal/model"
)

type ContentModerationRequest struct {
	AppID    int
	Scene    string
	TargetID string
	UserID   int64
	Content  string
}

type ContentModerationResult struct {
	Content string
	Blocked bool
	Words   []string
	Action  string
	Hits    []SensitiveHitSeed
}

type SensitiveHitSeed struct {
	Word   model.IMSensitiveWord
	Action string
}

func (r *ChatRepository) ModerateContent(ctx context.Context, req ContentModerationRequest) (*ContentModerationResult, error) {
	res := &ContentModerationResult{Content: req.Content}
	if !r.Ready() || strings.TrimSpace(req.Content) == "" {
		return res, nil
	}
	var words []model.IMSensitiveWord
	if err := r.db.WithContext(ctx).
		Where("status = ? AND (app_id = 0 OR app_id = ?)", model.SensitiveWordStatusEnabled, req.AppID).
		Order("app_id DESC, id ASC").
		Find(&words).Error; err != nil {
		return nil, err
	}
	for _, word := range words {
		if !sensitiveWordMatched(req.Content, word) {
			continue
		}
		res.Words = append(res.Words, word.Word)
		action := normalizeSensitiveAction(word.Action)
		if action == "reject" {
			res.Blocked = true
			res.Action = action
		}
		if action == "replace" {
			replacement := word.Replacement
			if replacement == "" {
				replacement = "***"
			}
			res.Content = replaceSensitiveWord(res.Content, word.Word, replacement, word.MatchType)
			if res.Action == "" {
				res.Action = action
			}
		}
		if action == "review" && res.Action == "" {
			res.Action = action
		}
		res.Hits = append(res.Hits, SensitiveHitSeed{Word: word, Action: action})
	}
	return res, nil
}

func (r *ChatRepository) RecordContentModerationHits(ctx context.Context, req ContentModerationRequest, res *ContentModerationResult, messageID uint64) error {
	if !r.Ready() || res == nil || len(res.Hits) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return recordContentModerationHits(tx, req, res, messageID)
	})
}

func recordContentModerationHits(tx *gorm.DB, req ContentModerationRequest, res *ContentModerationResult, messageID uint64) error {
	if res == nil || len(res.Hits) == 0 {
		return nil
	}
	for _, seed := range res.Hits {
		if err := recordSensitiveHit(tx, req, seed.Word, seed.Action, messageID); err != nil {
			return err
		}
	}
	return nil
}

func recordSensitiveHit(tx *gorm.DB, req ContentModerationRequest, word model.IMSensitiveWord, action string, messageID uint64) error {
	hit := model.IMSensitiveHit{
		AppID:          req.AppID,
		WordID:         word.ID,
		Word:           word.Word,
		Scene:          strings.TrimSpace(req.Scene),
		TargetID:       strings.TrimSpace(req.TargetID),
		MessageID:      messageID,
		UserID:         req.UserID,
		Action:         action,
		ContentSnippet: sensitiveSnippet(req.Content),
	}
	if err := tx.Create(&hit).Error; err != nil {
		return err
	}
	return tx.Model(&model.IMSensitiveWord{}).
		Where("id = ?", word.ID).
		UpdateColumn("hit_count", gorm.Expr("hit_count + ?", 1)).Error
}

func sensitiveWordMatched(content string, word model.IMSensitiveWord) bool {
	raw := strings.TrimSpace(word.Word)
	if raw == "" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(word.MatchType)) {
	case "exact":
		return strings.EqualFold(strings.TrimSpace(content), raw)
	default:
		return strings.Contains(strings.ToLower(content), strings.ToLower(raw))
	}
}

func normalizeSensitiveAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "replace":
		return "replace"
	case "review":
		return "review"
	default:
		return "reject"
	}
}

func replaceSensitiveWord(content string, raw string, replacement string, matchType string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return content
	}
	if strings.EqualFold(strings.TrimSpace(matchType), "exact") {
		if strings.EqualFold(strings.TrimSpace(content), raw) {
			return replacement
		}
		return content
	}
	out := content
	lowerNeedle := strings.ToLower(raw)
	for {
		lowerOut := strings.ToLower(out)
		idx := strings.Index(lowerOut, lowerNeedle)
		if idx < 0 {
			return out
		}
		out = out[:idx] + replacement + out[idx+len(raw):]
	}
}

func sensitiveSnippet(content string) string {
	runes := []rune(strings.TrimSpace(content))
	if len(runes) > 180 {
		return string(runes[:180])
	}
	return string(runes)
}
