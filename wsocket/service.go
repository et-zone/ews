package wsocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ConnPool struct {
	sync.Mutex
	Conns map[string]*Conn
}

var Conns = ConnPool{
	sync.Mutex{},
	map[string]*Conn{},
}

const (
	connKey = "Sec-Websocket-Key"
	ID      = "uid" //unique user id
	chSize  = 20
)

var CheckCliLive = 1 //second

func SetCheckCliLive(second int) {
	CheckCliLive = second
}

type Service struct {
	websocket.Upgrader
	Ping func(d string) error
	Pong func(d string) error
}

func NewService(ping,pong func(string)error)*Service{
	return &Service{websocket.Upgrader{},ping,pong}
}

func (s *Service) NewConn(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	c, err := s.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return nil, err
	}
	id := r.Header.Get(ID)
	if id == "" {
	}
	conn := &Conn{
		Conn:  c,
		Token: r.Header.Get(connKey),
		ID:    id,
	}
	Conns.Conns[id] = conn
	if s.Ping != nil {
		conn.Conn.SetPingHandler(s.Ping)
	}
	if s.Pong != nil {
		conn.Conn.SetPingHandler(s.Pong)
	}

	return conn, err
}

//if this server ping to dest and succ , this server will auth run Pong func
func Pong(d string) error {
	fmt.Println("client Pong succ!")
	return nil
}

type Conn struct {
	*websocket.Conn
	Token string
	ID    string
}

func (c *Conn) Run(worker Worker, checkCliLive bool) error {
	var err error
	var mt int
	var msg []byte
	ch := make(chan string, 0)
	defer func() {
		Conns.Lock()
		defer Conns.Unlock()
		c.Close()
		delete(Conns.Conns, c.ID)
		close(ch)
		worker=nil
		fmt.Println("succ del conn!",c.ID)
	}()

	if checkCliLive == true {
		go func(t int) {
			for { // Determine whether the client is alive or not, and exit the current connection if it is not
				err = c.WriteMessage(websocket.PongMessage, nil)
				if err != nil {
					if !strings.Contains(err.Error(),ClosedConn){
						log.Println("pong:", err)
					}
					if _, ok := <-ch;ok==true {
						ch <- c.ID
					}
					break
				}
				time.Sleep(time.Duration(t) * time.Second)
			}

		}(CheckCliLive)
	}
	for {
		mt, msg, err = c.ReadMessage()
		if err != nil {
			if !strings.Contains(err.Error(),ClosedConn){
				log.Println("read:", err, mt)
			}
			break
		}
		switch mt {
		case websocket.TextMessage:
			//err = c.WriteMessage(mt, message)
			//if err != nil {
			//	log.Println("write:", err)
			//	break
			//}

			err = worker.TextMessage(c, msg)
			if err != nil {
				if !strings.Contains(err.Error(),ClosedConn){
					log.Println("TextMessage err:", err)
				}
				return err
			}
		case websocket.BinaryMessage:
			//err = c.WriteMessage(mt, msg)
			//if err != nil {
			//	log.Println("write:", err)
			//	break
			//}
			err = worker.BinaryMessage(c, msg)
			if err != nil {
				if !strings.Contains(err.Error(),ClosedConn){
					log.Println("BinaryMessage err:", err)
				}
				return err
			}
		case websocket.CloseMessage:
			err = worker.CloseMessage(c)
			if err != nil {
				log.Println("CloseMessage err:", err)
			}
			return err

		default: //close err
			fmt.Println("default", mt, msg)
			return nil
		}

		select {
		case <-ch:
			return nil
		default:
		}
		log.Printf("recv: %s,mtype:%v", msg, mt)
	}
	return err
}

type Worker interface {
	TextMessage(conn *Conn, msg []byte) error
	BinaryMessage(conn *Conn, msg []byte) error
	CloseMessage(conn *Conn) error
}

func RecoverSrv(){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt,os.Kill,syscall.SIGHUP)
	select {
	case <-interrupt:
		return
	}
}

func RecoverCli(c *websocket.Conn,done chan struct{}){
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt,os.Kill,syscall.SIGHUP)
	select {
	case <-interrupt:
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseMessage, ""))
		if err != nil {
			log.Println("write close:", err)
			return
		}
		return
	case <-done:
		return
	}
}