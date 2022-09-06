
package main

import (
	"flag"
	"fmt"
	"github.com/et-zone/ews/worker"
	"github.com/et-zone/ews/wsocket"
	"log"
	"net/http"
)

var add = "localhost:8080"

var Service *wsocket.Service// use default options

func handler(w http.ResponseWriter, r *http.Request) {
	conn,_:=Service.NewConn(w,r)
	fmt.Println(conn.ID,conn.Token)
	wk := wsocket.Worker(&worker.WorkDmo{})
	conn.Run(wk,false)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", handler)
	//http.HandleFunc("/", home)
	go log.Fatal(http.ListenAndServe(add, nil))
	wsocket.RecoverSrv()
}

func init(){
	Service=wsocket.NewService(nil,nil)
	//Service=wsocket.NewService(nil,wsocket.Pong)
}

