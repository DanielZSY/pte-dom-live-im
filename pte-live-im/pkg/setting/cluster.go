package setting

import (
	"log"
	"os"
	"strconv"
	"strings"

	"pte_live_im/pkg/cluster"
)

type clusterConf struct {
	Enabled string `yaml:"enabled"` // auto | true | false
	PodHost string `yaml:"podHost"`
}

func normalizeCluster() {
	if strings.TrimSpace(clusterCfg.Enabled) == "" {
		clusterCfg.Enabled = "auto"
	}
	clusterCfg.Enabled = strings.ToLower(strings.TrimSpace(clusterCfg.Enabled))

	clusterOn := cluster.Enabled(clusterCfg.Enabled, EtcdSetting.Endpoints)
	if clusterOn {
		CommonSetting.Cluster = true
	}
	if host := cluster.PodHost(clusterCfg.PodHost); host != "" {
		GlobalSetting.LocalHost = host
	}

	log.Printf("cluster: enabled=%v mode=%s k8s=%v im.cluster=%v localHost=%s",
		clusterOn, clusterCfg.Enabled, cluster.InKubernetes(), CommonSetting.Cluster, GlobalSetting.LocalHost)
}

func normalizeRedisAddrs() {
	if len(RedisSetting.Addrs) == 0 && strings.TrimSpace(RedisSetting.Addr) != "" {
		RedisSetting.Addrs = []string{strings.TrimSpace(RedisSetting.Addr)}
	}
	addrs := make([]string, 0, len(RedisSetting.Addrs))
	for _, a := range RedisSetting.Addrs {
		if v := strings.TrimSpace(a); v != "" {
			addrs = append(addrs, v)
		}
	}
	RedisSetting.Addrs = addrs
	if len(RedisSetting.Addrs) > 0 && RedisSetting.Addr == "" {
		RedisSetting.Addr = RedisSetting.Addrs[0]
	}
	mode := strings.ToLower(strings.TrimSpace(RedisSetting.Mode))
	if mode == "" {
		if len(RedisSetting.Addrs) > 1 {
			mode = "cluster"
		} else {
			mode = "single"
		}
	}
	RedisSetting.Mode = mode

	if v := strings.TrimSpace(os.Getenv("REDIS_ADDR")); v != "" {
		RedisSetting.Addr = v
		if len(RedisSetting.Addrs) == 0 {
			RedisSetting.Addrs = []string{v}
		}
	}
	for i := 2; i <= 8; i++ {
		if v := strings.TrimSpace(os.Getenv("REDIS_ADDR_" + strconv.Itoa(i))); v != "" {
			RedisSetting.Addrs = append(RedisSetting.Addrs, v)
		}
	}
}

var clusterCfg clusterConf
