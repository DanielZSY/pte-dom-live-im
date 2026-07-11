package servers

import (
	"strings"

	"pte_live_im/define/livecode"
)

func SceneChannel(scene, roomID string) string {
	scene = strings.ToLower(strings.TrimSpace(scene))
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return ""
	}
	switch scene {
	case "shop":
		return livecode.GroupName(roomID)
	case "show", "voice":
		return scene + ":" + roomID
	default:
		return ""
	}
}
