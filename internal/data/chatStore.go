package data

import (
	"fmt"
	"goantisc/internal/core/sjson"
	"os"
	"sync"
)

type serializedChatInfo struct {
	Administrators []ChatMember `json:"administrators"`
	AllowMembers   []ChatMember `json:"allow_members"`
	AllowWords     []string     `json:"allow_words"`
	BlockWords     []string     `json:"block_words"`
}

func NewChatInfoStore() *ChatInfoStore {
	return &ChatInfoStore{
		data: map[int64]*ChatInfo{},
	}
}

type ChatInfoStore struct {
	data map[int64]*ChatInfo
	mu   sync.Mutex
}

func (s *ChatInfoStore) Save(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	serializedData := make(map[int64]serializedChatInfo)
	for id, chatInfo := range s.data {
		serializedData[id] = serializedChatInfo{
			Administrators: chatInfo.administrators,
			AllowMembers:   chatInfo.allowMembers,
			AllowWords:     chatInfo.GetAllowWords(),
			BlockWords:     chatInfo.GetBlockWords(),
		}
	}

	// 使用 sjson.Marshal 序列化
	jsonData, err := sjson.Marshal(serializedData)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	file, err := os.Create(path)
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

func (s *ChatInfoStore) Load(path string) error {
	file, err := os.Open(path)
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
