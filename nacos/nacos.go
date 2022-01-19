package nacos

import (
	"encoding/json"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// 初始化 Nacos
type InitNacosRequest struct {
	Host                string `json:"Host"`
	Port                uint64 `json:"Port"`
	NamespaceId         string `json:"NamespaceId"`
	DataId              string `json:"DataId"`
	Group               string `json:"Group"`
	NotLoadCacheAtStart bool   `json:"NotLoadCacheAtStart"`
	LogDir              string `json:"LogDir"`
	CacheDir            string `json:"CacheDir"`
	TimeoutMs           uint64 `json:"TimeoutMs"`  // (选填)
	RotateTime          string `json:"RotateTime"` // (选填)
	MaxAge              int64  `json:"MaxAge"`     // (选填)
	LogLevel            string `json:"LogLevel"`   // (选填)
}

func InitNacos(init InitNacosRequest, config interface{}) (config_client.IConfigClient, error) {
	// 初始化默认值
	if init.TimeoutMs == 0 {
		init.TimeoutMs = 5000
	}
	if init.RotateTime == "" {
		init.RotateTime = "1h"
	}
	if init.MaxAge == 0 {
		init.MaxAge = 3
	}
	if init.LogLevel == "" {
		init.LogLevel = "debug"
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: init.Host,
			Port:   init.Port,
		},
	}
	clientConfig := constant.ClientConfig{
		NamespaceId:         init.NamespaceId,
		TimeoutMs:           init.TimeoutMs,
		NotLoadCacheAtStart: init.NotLoadCacheAtStart,
		LogDir:              init.LogDir,
		CacheDir:            init.CacheDir,
		RotateTime:          init.RotateTime,
		MaxAge:              init.MaxAge,
		LogLevel:            init.LogLevel,
	}

	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("init nacos fail: %s", err.Error())
	}

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: init.DataId,
		Group:  init.Group,
	})
	if err != nil {
		return nil, fmt.Errorf("init nacos fail: %s", err.Error())
	}

	err = json.Unmarshal([]byte(content), &config)
	if err != nil {
		return nil, fmt.Errorf("init nacos fail: %s", err.Error())
	}

	return configClient, nil
}

// method 对返回错误进行错误处理 可以默认nil
func ListenNacos(init InitNacosRequest, client config_client.IConfigClient, config interface{}, method func(error)) {
	_ = client.ListenConfig(vo.ConfigParam{
		DataId: init.DataId,
		Group:  init.Group,
		OnChange: func(_, _, _, data string) {
			if method == nil {
				json.Unmarshal([]byte(data), &config)
			} else {
				err := json.Unmarshal([]byte(data), &config)
				method(err)
			}
		},
	})
}
