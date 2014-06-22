package main

import (
	"github.com/bnagy/alpcgo"
	"github.com/bnagy/w32"
	"log"
	"math/rand"
	"time"
)

const (
	srvName = "\\Sessions\\1\\BaseNamedObjects\\GoEchoSrv"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	clientMsg := w32.NewAlpcShortMessage()
	connmsg := "HELLO!"
	clientMsg.SetData([]byte(connmsg))

	log.Printf("Client1: Connection to %s as %s, connection message % x", srvName, "<unspecified>", connmsg)
	hClientComm, err := alpcgo.ConnectPort(srvName, "", &clientMsg)
	if err != nil {
		log.Fatalf("Failed to connect client1: %v", err)
	}
	log.Printf("Client1 Connected! Handle is 0x%X", hClientComm)

	clientMsg.Reset()

	for {

		time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
		msg := time.Now().String()

		// Reset the buffer, or the fields set in the previous recv will cause
		// the message to be rejected - new ALPC messages should have most
		// fields zeroed, as they are filled in by the kernel
		clientMsg.Reset()
		clientMsg.SetData([]byte(msg))

		log.Printf("Client: Sending %s to handle %x", msg, hClientComm)
		err := w32.NtAlpcSendWaitReceivePort(
			hClientComm,
			0,
			&clientMsg,
			nil,
			&clientMsg,
			nil,
			nil,
			nil,
		)
		if err != nil {
			log.Fatalf("Client: Recv Error: %v", err)
		}
		log.Printf("Client: Response Type: 0x%x Data: %s ", clientMsg.Type, string(clientMsg.GetData()))
	}

}
