package utils

import "github.com/bytedance/sonic"

func Marshal(v any) (string, error) {
	return sonic.Marshal(v)
}

func Unmarshal(data string, v any) error {
	return sonic.Unmarshal(data, v)
}
