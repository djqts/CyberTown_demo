package event

import "encoding/json"

// Marshal 将事件编码为 JSON 字节。
func Marshal(e *Event) ([]byte, error) {
	return json.Marshal(e)
}

// Unmarshal 从 JSON 字节解码为事件。
func Unmarshal(data []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
