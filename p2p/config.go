package p2p

import (
	"encoding/json"
	"os"
)

const (
	BootstrapInfoFile = "bootstrap.json"
)

type BootstrapInfo struct {
	MultiAddr string `json:"multi_addr"`
}

func SaveBootstrapInfo(addr string) error {
	info := BootstrapInfo{
		MultiAddr: addr,
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(BootstrapInfoFile, data, 0644)
}

func LoadBootstrapInfo() (string, error) {
	data, err := os.ReadFile(BootstrapInfoFile)
	if err != nil {
		return "", err
	}
	var info BootstrapInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	}
	return info.MultiAddr, nil
}
