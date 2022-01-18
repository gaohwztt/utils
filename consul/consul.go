package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
)

type InitConsulRequest struct {
	Name                           string
	Tags                           []string
	Timeout                        int
	Interval                       int
	DeregisterCriticalServiceAfter int
	CheckRouter                    string // 健康检查路由
	IsSSL                          bool   // 是否是https
	Consul                         struct {
		Host string
		Port int
	}
	Project struct {
		Host string
		Port int
	}
}

// consul 初始化
func InitConsul(init InitConsulRequest) (*api.Client, string, error) {
	cfg := api.DefaultConfig()

	// 连接consul
	cfg.Address = fmt.Sprintf("%s:%d", init.Consul.Host, init.Consul.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, "", fmt.Errorf("consul init fail, please check consul host and port is right")
	}

	// 生成 uuid
	serverID := uuid.NewV4().String()
	// 是否是 https
	initHttp := "http"
	if init.IsSSL {
		initHttp = "https"
	}

	// 服务注册
	err = client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		Name:    init.Name,
		ID:      serverID,
		Address: init.Project.Host,
		Port:    init.Project.Port,
		Tags:    init.Tags,
		Check: &api.AgentServiceCheck{
			HTTP: fmt.Sprintf("%s://%s:%d/%s", initHttp, init.Project.Host, init.Project.Port, init.CheckRouter),
		},
	})

	if err != nil {
		return nil, "", fmt.Errorf("consul service register fail, please check parameter is right")
	}

	return client, serverID, nil
}
