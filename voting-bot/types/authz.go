package types

import "time"

type Grants struct {
	Grants []Grant `json:"grants"`
}

type Grant struct {
	Authorization struct {
		Type string `json:"@type"`
		Msg  string `json:"msg"`
	} `json:"authorization"`
	Expiration time.Time `json:"expiration"`
}
