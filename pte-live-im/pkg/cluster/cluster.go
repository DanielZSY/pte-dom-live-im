package cluster

import (
	"os"
	"strings"
)

func Enabled(mode string, etcdEndpoints []string) bool {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		if strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST")) != "" {
			return true
		}
		for _, ep := range etcdEndpoints {
			if strings.TrimSpace(ep) != "" {
				return true
			}
		}
		return false
	}
}

func InKubernetes() bool {
	return strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST")) != ""
}

func PodHost(configured string) string {
	if v := strings.TrimSpace(configured); v != "" {
		return v
	}
	for _, key := range []string{"POD_IP", "HOST_IP", "IM_LOCAL_HOST"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}
