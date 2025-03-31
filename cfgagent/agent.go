package cfgagent

import (
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type cfgState struct {
	item   model.ConfigItem
	inUsed bool
}

type ConfigClient struct {
	Username    string
	Password    string
	ServerAddr  string
	NamespaceID string
	GroupID     string
	Number      int

	client config_client.IConfigClient
	oldCfg map[string]*cfgState
}

func (cfg *ConfigClient) NewClient() error {
	// 创建客户端配置
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(cfg.NamespaceID),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithUsername(cfg.Username),
		constant.WithPassword(cfg.Password),
	)

	// 创建服务端配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: cfg.ServerAddr,
			Port:   8848,
		},
	}

	// 创建配置客户端
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create config client: %w", err)
	}

	cfg.client = configClient
	cfg.oldCfg = make(map[string]*cfgState)
	return nil
}

// 封装函数用于检查配置是否更新
func isConfigUpdated(newItem, oldItem *cfgState) bool {
	return newItem.item.Md5 != oldItem.item.Md5
}

// 封装函数用于更新日志记录
func updateLogEntry(log map[string]string, action, dataID string) {
	log[action] = strings.TrimSpace(fmt.Sprintf("%s %s", log[action], dataID))
}

func (cfg *ConfigClient) LoadAllConfig() (map[string]string, error) {
	// 列出指定命名空间下的所有配置
	listConfig, err := cfg.client.SearchConfig(vo.SearchConfigParam{
		Search:   "blur",
		PageNo:   1,
		PageSize: cfg.Number,
		DataId:   "",
		Group:    cfg.GroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search config: %w", err)
	}

	newCfg := make(map[string]*cfgState)
	for _, item := range listConfig.PageItems {
		newCfg[item.DataId] = &cfgState{
			item:   item,
			inUsed: false,
		}
	}

	updateLog := make(map[string]string, 0)

	// 处理更新和新增
	for dataID, newItem := range newCfg {
		if oldItem, exists := cfg.oldCfg[dataID]; exists {
			if isConfigUpdated(newItem, oldItem) {
				// 更新操作
				updateLogEntry(updateLog, "update", dataID)
			}
		} else {
			// 新增操作
			updateLogEntry(updateLog, "add", dataID)
		}
	}

	// 处理删除
	for dataID := range cfg.oldCfg {
		if _, exists := newCfg[dataID]; !exists {
			// 删除操作
			updateLogEntry(updateLog, "delete", dataID)
		}
	}

	// 更新旧配置
	cfg.oldCfg = newCfg

	return updateLog, nil
}
