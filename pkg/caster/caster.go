package caster

import "encoding/json"

type ChannelCaster[T any] interface {
	From(string) (T, error)
	To(T) (string, error)
}

type JSONChannelCaster[T any] struct{}

func (jc JSONChannelCaster[T]) From(data string) (T, error) {
	var v T
	err := json.Unmarshal([]byte(data), &v)
	return v, err
}

func (jc JSONChannelCaster[T]) To(v T) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}
