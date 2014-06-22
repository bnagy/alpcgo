package main

import (
	"fmt"
	"github.com/bnagy/w32"
	"log"
	"math/rand"
	"net"
	"net/rpc/jsonrpc"
	"time"
)

type ConnReq struct {
	Port string
	Msg  []byte
}

type Message struct {
	Handle w32.HANDLE
	Flags  uint32
	w32.PORT_MESSAGE
	Payload    []byte
	Attributes []byte
}

func fill(sm *w32.AlpcShortMessage, rpcm *Message) {
	rpcm.PORT_MESSAGE = sm.PORT_MESSAGE
	rpcm.Payload = sm.GetData()
}

func main() {

	client, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	rpc := jsonrpc.NewClient(client)

	// Synchronous call
	args := &ConnReq{
		"\\Sessions\\1\\BaseNamedObjects\\GoEchoSrv",
		[]byte("A package from the Transvaal! How strange!"),
	}
	var h w32.HANDLE

	err = rpc.Call("ALPC.Connect", args, &h)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer rpc.Call("ALPC.Close", &h, nil)

	fmt.Printf("Connected! Handle is: 0x%X", h)
	clientMsg := w32.NewAlpcShortMessage()
	var rpcSend, rpcRecv Message
	rpcSend.Handle = h

	for {

		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		msg := time.Now().String()

		// Reset the buffer, or the fields set in the previous recv will cause
		// the message to be rejected - new ALPC messages should have most
		// fields zeroed, as they are filled in by the kernel
		clientMsg.Reset()
		clientMsg.SetData([]byte(msg))
		fill(&clientMsg, &rpcSend)

		log.Printf("sending %s to handle %x", msg, h)
		err = rpc.Call("ALPC.Send", &rpcSend, &rpcRecv)
		if err != nil {
			log.Fatalf("recv Error: %v", err)
		}

		log.Printf("response Type: 0x%x Data: [% x] ", rpcRecv.Type, rpcRecv.Payload)
	}

}
