package worker

import (
	"fmt"
	"ews/wsocket"
)
// you can realization something what you want to do， all of business can complete with it

// it is a daemon
type WorkDmo struct {}

func (w *WorkDmo)TextMessage(conn *wsocket.Conn,msg []byte) error{
	fmt.Println(string(msg))
	return nil
}

func (w *WorkDmo)BinaryMessage(conn *wsocket.Conn,msg []byte) error{
	fmt.Println(string(msg))
	return nil
}
func (w *WorkDmo)CloseMessage(conn *wsocket.Conn) error{
	return nil
}