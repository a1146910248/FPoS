package config

import (
	"FPoS/core/ethereum"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Config struct {
	Ethereum *ethereum.EthereumConfig `yaml:"ethereum"`
}

var defaultConfig = &Config{
	Ethereum: &ethereum.EthereumConfig{
		RPCURL:        "http://localhost:8552",
		GasLimit:      3000000,
		GasPrice:      20000000000,
		ConfirmBlocks: 2,
	},
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	config := defaultConfig

	// 如果配置文件存在，则读取
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// 环境变量覆盖
	if url := os.Getenv("ETH_RPC_URL"); url != "" {
		config.Ethereum.RPCURL = url
	}
	if addr := os.Getenv("L2_CONTRACT_ADDR"); addr != "" {
		config.Ethereum.ContractAddress = addr
	}
	if key := os.Getenv("ETH_PRIVATE_KEY"); key != "" {
		config.Ethereum.PrivateKey = key
	}

	return config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
