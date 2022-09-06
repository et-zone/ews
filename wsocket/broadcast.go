package wsocket

import (
	"fmt"
	"github.com/gorilla/websocket"
)

func SendBroadcast(msg []byte){
	for _,conn:=range Conns.Conns{
		err:=conn.WriteMessage(websocket.BinaryMessage,msg)
		if err!=nil{
			fmt.Println(err)
		}
	}
}

