package core

import "github.com/pelletier/go-toml"

type ServiceConfig struct {
	BotToken string `toml:"bot_token"`
}

func LoadConfig(filePath string) (*ServiceConfig, error) {
	var ret ServiceConfig
	tree, err := toml.LoadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := tree.Unmarshal(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}
