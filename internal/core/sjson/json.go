package sjson

import "github.com/bytedance/sonic"

var (
	Marshal         = sonic.Marshal
	MarshalIndent   = sonic.MarshalIndent
	MarshalString   = sonic.MarshalString
	Unmarshal       = sonic.Unmarshal
	UnmarshalString = sonic.UnmarshalString
)
