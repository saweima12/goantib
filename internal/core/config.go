package core

import (
	"github.com/pelletier/go-toml"
)

type storedServiceConfig struct {
	BotToken          string   `toml:"bot_token"`
	DefaultAllowWords []string `toml:"default_allow_words"`
}

type ServiceConfig struct {
	BotToken          string
	DefaultAllowWords map[string]struct{}
}

func LoadConfig(filePath string) (*ServiceConfig, error) {
	var temp storedServiceConfig
	tree, err := toml.LoadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err := tree.Unmarshal(&temp); err != nil {
		return nil, err
	}

	ret := ServiceConfig{
		BotToken:          temp.BotToken,
		DefaultAllowWords: make(map[string]struct{}),
	}

	for _, word := range temp.DefaultAllowWords {
		ret.DefaultAllowWords[word] = struct{}{}
	}

	return &ret, nil
}
