package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var addr = flag.String("addr", "0.0.0.0:8000", "http service address")

var upgradeHelper = websocket.Upgrader{}

var streamNameMapRolling = make(map[string]int)
var streamNameMapStale = make(map[string]int)
var staleMutex = &sync.RWMutex{}
var rollingMutex = &sync.RWMutex{}

func ping(w http.ResponseWriter, r *http.Request) {
	c, err := upgradeHelper.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgradeErr:", r.Host, " ", err)
		return
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Println("WSCloseErr:", err)
		}
	}(c)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			//log.Println("readErr:", err) // not needed, triggers every page close
			break
		}
		rollingMutex.Lock()
		streamNameMapRolling[string(message)]++
		rollingMutex.Unlock()
	}
}

func data(w http.ResponseWriter, _ *http.Request) {
	staleMutex.RLock()
	jsonRes, err := json.Marshal(streamNameMapStale)
	staleMutex.RUnlock()
	if err != nil {
		log.Println("readErr:", err)
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonRes)
	if err != nil {
		log.Println("sendErr:", err)
	}
}

func home(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("Hello there!"))
}

func periodicWiper() {
	for range time.Tick(time.Second * 10) {
		staleMutex.Lock()
		streamNameMapStale = make(map[string]int)
		rollingMutex.Lock()
		for k, v := range streamNameMapRolling {
			streamNameMapStale[k] = v
		}
		staleMutex.Unlock()
		streamNameMapRolling = make(map[string]int)
		rollingMutex.Unlock()
	}
}

func main() {
	log.Println("Server started!")
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/data", data)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/", home)
	go periodicWiper()
	log.Fatal(http.ListenAndServe(*addr, nil))
}
