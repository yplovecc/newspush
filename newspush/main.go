package main

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/golang/glog"
	"github.com/robfig/cron"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	badgerDir = flag.String("badgerdir", "badger", "badger db dir")
	confDir   = flag.String("confdir", "conf.json", "conf dir")
)

func stop(sigs chan os.Signal, exitCh chan int) {
	<-sigs
	glog.Info("recieve stop signal")
	close(exitCh)
}

func startCron(wg *sync.WaitGroup, exitCh chan int, re chan os.Signal, kv *badger.KV) {
	defer wg.Done()
	for {
		pushitems, err := LoadConfig(*confDir)
		if err != nil {
			return
		}
		c := cron.New()
		c.Start()
		for _, item := range pushitems {
			spec := fmt.Sprintf("0 %d %d * * *", item.Minute, item.Hour)
			c.AddJob(spec, PushCronJob{kv, item})
		}
		select {
		case <-exitCh:
			c.Stop()
			return
		case <-re:
			glog.Info("reload config; restart crontab")
			c.Stop()
			break
		}
	}
}

func main() {
	flag.Parse()
	glog.Info("server start")
	defer glog.Info("server exit...")
	kv, err := NewBadger(*badgerDir)
	if err != nil {
		glog.Error("kv store start failed, exit: %s", err)
		return
	}
	defer kv.Close()

	exitCh := make(chan int)
	sigs := make(chan os.Signal)
	restartSig := make(chan os.Signal)
	signal.Notify(restartSig, syscall.SIGUSR1)
	var wg sync.WaitGroup
	wg.Add(1)
	go startCron(&wg, exitCh, restartSig, kv)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go stop(sigs, exitCh)
	wg.Wait()
}
