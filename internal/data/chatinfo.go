package data

import (
	"sync"
	"time"
)

type ChatMember struct {
	Id   int64
	Name string
}

type ChatInfo struct {
	administrators []ChatMember
	allowMembers   []ChatMember
	allowWords     map[string]struct{}
	blockWords     map[string]struct{}

	UpdateAt time.Time
	mu       sync.Mutex
}

func NewChatInfo() *ChatInfo {
	return &ChatInfo{
		administrators: []ChatMember{},
		allowMembers:   make([]ChatMember, 0),
		allowWords:     make(map[string]struct{}),
		blockWords:     make(map[string]struct{}),
		mu:             sync.Mutex{},
	}
}

func (ci *ChatInfo) Update(f func(info *ChatInfo) error) error {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	return f(ci)
}

func (ci *ChatInfo) GetAdministrators() []ChatMember {
	ci.mu.Lock()
	defer ci.mu.Unlock()

	ret := make([]ChatMember, len(ci.administrators))
	copy(ret, ci.administrators)
	return ret
}

func (ci *ChatInfo) GetAllowMembers() []ChatMember {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make([]ChatMember, len(ci.allowMembers))
	copy(ret, ci.allowMembers)
	return ret
}

func (ci *ChatInfo) GetAllowWords() map[string]struct{} {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make(map[string]struct{})
	for word := range ci.allowWords {
		ret[word] = struct{}{}
	}
	return ret
}

func (ci *ChatInfo) GetBlockWords() map[string]struct{} {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make(map[string]struct{})
	for word := range ci.blockWords {
		ret[word] = struct{}{}
	}
	return ret
}

func (ci *ChatInfo) GetAllowWordList() []string {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make([]string, 0, len(ci.allowWords))
	for word := range ci.allowWords {
		ret = append(ret, word)
	}
	return ret
}

func (ci *ChatInfo) GetBlockWordList() []string {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make([]string, 0, len(ci.allowWords))
	for word := range ci.blockWords {
		ret = append(ret, word)
	}
	return ret
}

func (ci *ChatInfo) AddAllowWord(word string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.allowWords[word] = struct{}{}
}

func (ci *ChatInfo) AddBlockWord(word string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.blockWords[word] = struct{}{}
}

func (ci *ChatInfo) RemoveAllowWord(word string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	delete(ci.allowWords, word)
}

func (ci *ChatInfo) RemoveBlockWord(word string) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	delete(ci.blockWords, word)
}

func (ci *ChatInfo) UpdateAdministrators(administrators []ChatMember) {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.administrators = administrators
}
