package queue

import (
	"testing"

	"pte_live_im/define/livecode"
	"pte_live_im/servers/live"
)

func TestDispatchGiftAndLinkMicApplyAreSeparated(t *testing.T) {
	resetProcessedForTest()

	appId := "app-consumer-gift"
	roomId := "room-consumer-gift"
	userId := "user-consumer-gift"

	err := dispatch(Message{
		MessageId: "gift-message-1",
		AppId:     appId,
		RoomId:    roomId,
		UserId:    userId,
		Code:      livecode.GiftSend,
		Msg:       "gift",
		Data:      `{"giftId":"g1","giftName":"rocket","count":2,"amount":3.5}`,
	})
	if err != nil {
		t.Fatalf("dispatch gift: %v", err)
	}

	gifts, giftTotal := live.GiftList(appId, roomId, 1, 20)
	if giftTotal != 1 || len(gifts) != 1 {
		t.Fatalf("gift should be recorded once, got total=%d len=%d", giftTotal, len(gifts))
	}
	if gifts[0].UserId != userId {
		t.Fatalf("gift userId mismatch: %q", gifts[0].UserId)
	}
	if links := live.LinkMicList(appId, roomId); len(links) != 0 {
		t.Fatalf("gift must not create link-mic apply, got %d", len(links))
	}

	err = dispatch(Message{
		MessageId: "gift-message-1",
		AppId:     appId,
		RoomId:    roomId,
		UserId:    userId,
		Code:      livecode.GiftSend,
		Msg:       "gift",
		Data:      `{"giftId":"g1","giftName":"rocket","count":2,"amount":3.5}`,
	})
	if err != nil {
		t.Fatalf("dispatch duplicate gift: %v", err)
	}
	_, giftTotal = live.GiftList(appId, roomId, 1, 20)
	if giftTotal != 1 {
		t.Fatalf("duplicate message should be ignored, got gift total=%d", giftTotal)
	}

	err = dispatch(Message{
		MessageId: "linkmic-message-1",
		AppId:     appId,
		RoomId:    roomId,
		UserId:    userId,
		Code:      livecode.LinkMicApply,
		Msg:       "linkmic",
		Data:      `{"nick":"tester","avatar":"avatar.png"}`,
	})
	if err != nil {
		t.Fatalf("dispatch link-mic apply: %v", err)
	}

	links := live.LinkMicList(appId, roomId)
	if len(links) != 1 {
		t.Fatalf("link-mic apply should be recorded once, got %d", len(links))
	}
	if links[0].UserId != userId {
		t.Fatalf("link-mic userId mismatch: %q", links[0].UserId)
	}
}

func resetProcessedForTest() {
	processedIdsMu.Lock()
	defer processedIdsMu.Unlock()
	processedIds = make(map[string]struct{})
	processedIdOrder = nil
}
