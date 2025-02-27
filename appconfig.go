package gosms

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"strings"
)

// Settings 定义结构体来映射 TOML 配置文件
type Settings struct {
	ServerHost     string `toml:"serverHost"`
	ServerPort     int    `toml:"serverPort"`
	Retries        int    `toml:"retries"`
	BufferSize     int    `toml:"bufferSize"`
	BufferLow      int    `toml:"bufferLow"`
	MsgTimeout     int    `toml:"msgTimeout"`
	MsgCountOut    int    `toml:"msgCountOut"`
	MsgTimeoutLong int    `toml:"msgTimeoutLong"`
}

type Device struct {
	ComPort  string `toml:"comPort"`
	BaudRate int    `toml:"baudRate"`
}

type Config struct {
	Settings Settings          `toml:"settings"`
	Devices  map[string]Device `toml:"device"`
}

// GetConfig 加载并返回配置
func GetConfig(configFilePath string) (*Config, error) {
	var appConfig Config

	// 解析 TOML 配置文件
	_, err := toml.DecodeFile(configFilePath, &appConfig)
	if err != nil {
		return nil, err
	}

	// 测试配置是否有效
	ok, err := testConfig(appConfig)
	if !ok {
		return nil, err
	}

	return &appConfig, nil
}

// testConfig 检查必需的配置项
func testConfig(appConfig Config) (bool, error) {
	// 检查 SETTINGS 部分
	requiredFields := []struct {
		field string
		value interface{}
	}{
		{"ServerHost", appConfig.Settings.ServerHost},
		{"ServerPort", appConfig.Settings.ServerPort},
		{"Retries", appConfig.Settings.Retries},
		{"BufferSize", appConfig.Settings.BufferSize},
		{"BufferLow", appConfig.Settings.BufferLow},
		{"MsgTimeout", appConfig.Settings.MsgTimeout},
		{"MsgCountOut", appConfig.Settings.MsgCountOut},
		{"MsgTimeoutLong", appConfig.Settings.MsgTimeoutLong},
	}

	for _, field := range requiredFields {
		if field.value == nil || (field.field == "ServerHost" && strings.TrimSpace(field.value.(string)) == "") {
			return false, errors.New(fmt.Sprintf("Fatal: %s is not set", field.field))
		}
	}

	for key, device := range appConfig.Devices {
		if strings.TrimSpace(device.ComPort) == "" || device.BaudRate == 0 {
			return false, fmt.Errorf("fatal: Device %s configuration is incomplete", key)
		}
	}

	return true, nil
}
