package message

import (
	_ "embed"
	"encoding/json"

	"github.com/labstack/gommon/log"
)

//go:embed messages.json
var msgJSON []byte

var msgs map[string]string

// Read メッセージ一覧を読み込む
func Read() {
	if err := json.Unmarshal(msgJSON, &msgs); err != nil {
		panic("Cannot read messages.json")
	}
}

// Get keyからメッセージを取得する(keyがなければ空を返す)
func Get(key string) *string {
	msg, exists := msgs[key]
	if !exists {
		log.Errorf("Cannnot find this message key: %s", key)
	}
	return &msg
}
