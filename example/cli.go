package main

import (
	"flag"
	"fmt"
	"github.com/et-zone/ews/wsocket"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = "localhost:8080"

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	h:=http.Header{}
	h.Add("id","10001")
	c, _, err := websocket.DefaultDialer.Dial(u.String(),h)
	if err != nil {
		log.Fatal("dial:", err)
	}

	c.SetPongHandler(func (d string)error{
		fmt.Println("srv pong succ!")
		return nil
	})
	defer c.Close()

	done := make(chan struct{})

	go func() {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				if !strings.Contains(err.Error(),wsocket.ClosedConn){
					log.Println("read:", err)
				}
				if _,ok:=<-done;ok==true{
					done<- struct{}{}
				}
				break
			}
			log.Printf("recv: %s,%v", message,mt)
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, []byte(time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		if !strings.Contains(err.Error(),wsocket.ClosedConn){
			log.Println("write:", err)
		}
		return
	}

	go func() {
		for{//判断服务端是否存活，如果不存活就退出客户端
			err = c.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				if !strings.Contains(err.Error(),wsocket.ClosedConn){
					log.Println("ping:", err)
				}
				if _,ok:=<-done;ok==true{
					done<- struct{}{}
				}
				break
			}
			time.Sleep(time.Second)
		}
	}()
	wsocket.RecoverCli(c,done)
}