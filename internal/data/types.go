package data

import "sync"

type ChatMember struct {
	Id   int64
	Name string
}

type ChatInfo struct {
	administrators []ChatMember
	allowMembers   []ChatMember
	allowWords     map[string]struct{}
	blockWords     map[string]struct{}

	mu sync.Mutex
}

func NewChatInfo(administrator []ChatMember) *ChatInfo {
	return &ChatInfo{
		administrators: administrator,
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

func (ci *ChatInfo) GetAllowWords() []string {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make([]string, 0, len(ci.allowWords))
	for word := range ci.allowWords {
		ret = append(ret, word)
	}
	return ret
}

func (ci *ChatInfo) GetBlockWords() []string {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ret := make([]string, 0, len(ci.blockWords))
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
