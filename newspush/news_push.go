package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"github.com/crawlerclub/dl"
	"github.com/dgraph-io/badger"
	"github.com/golang/glog"
	"net/http"
	"os"
	"strings"
)

var (
	listUrl    = flag.String("listurl", "http://going.news/api/list?start=0&count=100", "news list url")
	contentUrl = flag.String("contenturl", "http://going.news/api/?url=", "news content url")
	pushSource = flag.String("pushsource", "https://fcm.googleapis.com/fcm/send", "push source url")
	toPrefix   = flag.String("toprefix", "/topics/HOT_NEWS_", "the prefix of push topic")
)

func NewBadger(dir string) (kv *badger.KV, err error) {
	glog.Info("kv store:badger starting")
	opt := badger.DefaultOptions
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return
		}
	}
	opt.Dir = dir
	opt.ValueDir = dir
	kv, err = badger.NewKV(&opt)
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func LoadConfig(path string) (pushlist []PushItem, err error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		glog.Error(err)
		return
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pushlist)
	glog.Infof("load push list: %s", pushlist)
	return
}

func Push(topic string, da PushData, key string) (st string, err error) {
	payload := PayLoad{To: topic, Data: da}
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		glog.Errorf("marshal error: %s, %s", payload, err)
		return
	}
	req, err := http.NewRequest("POST", *pushSource, bytes.NewBuffer(jsonStr))
	if err != nil {
		glog.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+key)
	client := &http.Client{}
	resp, err := client.Do(req)
	st = resp.Status
	return
}

func PushCron(kv *badger.KV, item PushItem) {
	glog.Info("crontab start")
	defer glog.Info("crontab end")
	da, err := GetPushData(kv)
	if err != nil {
		return
	}
	r, err := Push(*toPrefix+strings.ToUpper(item.Co), da, item.Key)
	if err != nil {
		glog.Error(err)
		return
	}
	glog.Infof("push to %s, key: %s, stauts: %s, data: %s", item.Co, item.Key, r, da)
	return
}

func GetPushData(kv *badger.KV) (da PushData, err error) {
	req := &dl.HttpRequest{Url: *listUrl,
		Method: "GET", UseProxy: false, Platform: "pc"}
	resp := dl.Download(req)
	if resp.Error != nil {
		glog.Error(resp.Error)
		err = resp.Error
		return
	}
	var list ListItems
	if err = json.Unmarshal(resp.Content, &list); err != nil {
		glog.Error(err)
		return
	}

	var pushurl string
	for _, item := range *list.List {
		if isexist, _ := kv.Exists([]byte(item.Url)); !isexist {
			pushurl = item.Url
			break
		}
	}
	if pushurl == "" {
		glog.Warning("no data to push")
		err = errors.New("no data to push")
		return
	}

	req1 := &dl.HttpRequest{Url: *contentUrl + pushurl,
		Method: "GET", UseProxy: false, Platform: "pc"}
	resp1 := dl.Download(req1)
	if resp1.Error != nil {
		glog.Error(resp1.Error)
		err = resp1.Error
		return
	}

	var inforec InfoRec
	err = json.Unmarshal(resp1.Content, &inforec)
	info := *inforec.Info
	da = PushData{Title: info.Title[:Min(len(info.Title), 80)],
		Summary: info.Text[:Min(len(info.Text), 140)],
		Url:     info.Url, Type: "NewsSdk"}
	glog.Infof("get push url : %s", da.Url)

	kv.Set([]byte(pushurl), []byte("1"), 0x00)
	return
}
