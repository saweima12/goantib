package data

import (
	"fmt"
	"goantisc/internal/core/sjson"
	"os"
	"sync"
	"time"
)

type ChatInfoUpdateFunc func(info *ChatInfo) error

type serializedChatInfo struct {
	Administrators []ChatMember `json:"administrators"`
	AllowMembers   []ChatMember `json:"allow_members"`
	AllowWords     []string     `json:"allow_words"`
	BlockWords     []string     `json:"block_words"`
}

func NewChatInfoStore(path string) *ChatInfoStore {
	return &ChatInfoStore{
		data: make(map[int64]*ChatInfo),
		mu:   sync.Mutex{},
		path: path,
	}
}

type ChatInfoStore struct {
	data map[int64]*ChatInfo
	path string
	mu   sync.Mutex
}

func (s *ChatInfoStore) GetChatInfo(chatId int64) *ChatInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	if origin, ok := s.data[chatId]; ok {
		ret := ChatInfo{
			administrators: origin.GetAdministrators(),
			allowMembers:   origin.GetAllowMembers(),
			allowWords:     origin.GetAllowWords(),
			blockWords:     origin.GetBlockWords(),
		}
		return &ret
	}

	ret := ChatInfo{
		administrators: make([]ChatMember, 0),
		allowMembers:   make([]ChatMember, 0),
		allowWords:     make(map[string]struct{}),
		blockWords:     map[string]struct{}{},
		UpdateAt:       time.Date(1970, 0, 1, 0, 0, 0, 0, time.Local),
	}
	return &ret
}

func (s *ChatInfoStore) Update(chatId int64, operateFunc ChatInfoUpdateFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, ok := s.data[chatId]; ok {
		operateFunc(item)
		return
	}

	s.data[chatId] = NewChatInfo()
	operateFunc(s.data[chatId])
}

func (s *ChatInfoStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	serializedData := make(map[int64]serializedChatInfo)
	for id, chatInfo := range s.data {
		serializedData[id] = serializedChatInfo{
			Administrators: chatInfo.administrators,
			AllowMembers:   chatInfo.allowMembers,
			AllowWords:     chatInfo.GetAllowWordList(),
			BlockWords:     chatInfo.GetBlockWordList(),
		}
	}

	jsonData, err := sjson.Marshal(serializedData)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	file, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %w", err)
	}

	return nil
}

func (s *ChatInfoStore) Load() error {
	file, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	buffer := make([]byte, fileStat.Size())

	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	serializedData := make(map[int64]serializedChatInfo)

	err = sjson.Unmarshal(buffer, &serializedData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[int64]*ChatInfo)
	for id, serializedChatInfo := range serializedData {
		chatInfo := ChatInfo{
			administrators: serializedChatInfo.Administrators,
			allowMembers:   serializedChatInfo.AllowMembers,
			allowWords:     make(map[string]struct{}),
			blockWords:     make(map[string]struct{}),
		}

		for _, word := range serializedChatInfo.AllowWords {
			chatInfo.allowWords[word] = struct{}{}
		}
		for _, word := range serializedChatInfo.BlockWords {
			chatInfo.blockWords[word] = struct{}{}
		}

		s.data[id] = &chatInfo
	}

	return nil
}
