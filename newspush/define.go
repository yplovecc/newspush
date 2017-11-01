package main

import (
	"github.com/crawlerclub/ce"
	"github.com/dgraph-io/badger"
)

type ListItem struct {
	Url       string `json:"url"`
	Image     string `json:"image"`
	Title     string `json:"title"`
	PublishTS int32  `json:"publish_ts"`
}

type ListItems struct {
	More bool
	List *[]ListItem
}

type NewsInfo struct {
	*ce.Doc
	PublishTS int32 `json:"publish_ts"`
}

type InfoRec struct {
	Info *NewsInfo
	Rec  *[]ListItem
}
type PushData struct {
	Title   string `json:"newsTitle"`
	Summary string `json:"newsSummary"`
	Url     string `json:"newsUrl"`
	Type    string `json:"type"`
}

type PayLoad struct {
	To   string   `json:"to"`
	Data PushData `json:"data"`
}

type PushItem struct {
	Co     string `json:"co"`
	Key    string `json:"key"`
	Hour   int32  `json:"hour"`
	Minute int32  `json:"minute"`
}

type PushCronJob struct {
	kv   *badger.KV
	item PushItem
}

func (p PushCronJob) Run() {
	PushCron(p.kv, p.item)
}
