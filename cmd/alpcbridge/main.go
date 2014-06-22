package main

import (
	"fmt"
	"github.com/bnagy/alpcgo"
	"github.com/bnagy/w32"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"unsafe"
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

type ALPC struct{}

var handles = make(map[w32.HANDLE]*w32.AlpcShortMessage)

func (alpc *ALPC) Connect(cr *ConnReq, reply *w32.HANDLE) error {

	clientMsg := w32.NewAlpcShortMessage()
	clientMsg.SetData(cr.Msg)

	hClientComm, err := alpcgo.ConnectPort(cr.Port, "", &clientMsg)
	if err != nil {
		log.Printf("error %v in Connect", err)
		return err
	}

	// Save a pointer to this client message. Any further traffic sent to this
	// client communication port can reuse the same message buffer. Probably
	// not all that threadsafe, but saves a lot of time and memory for
	// 'normal' use. 64k is 64k, after all.
	log.Printf("client connected! Handle is 0x%X", hClientComm)
	handles[hClientComm] = &clientMsg

	*reply = hClientComm
	return nil

}

func (alpc *ALPC) Close(h *w32.HANDLE, reply *int) error {
	delete(handles, *h)
	log.Printf("closed handle 0x%x", *h)
	return nil
}

func (alpc *ALPC) Send(m *Message, reply *Message) error {

	msgBuf, found := handles[m.Handle]
	if !found {
		log.Printf("invalid handle 0x%X in Send", m.Handle)
		return fmt.Errorf("invalid handle 0x%X", m.Handle)
	}

	log.Printf("sending [% x] to handle 0x%X", m.Payload, m.Handle)
	var pSendAttrs *w32.ALPC_MESSAGE_ATTRIBUTES
	if len(m.Attributes) > 0 {
		pSendAttrs = (*w32.ALPC_MESSAGE_ATTRIBUTES)(unsafe.Pointer(&m.Attributes[0]))
	}
	msgBuf.PORT_MESSAGE = m.PORT_MESSAGE
	msgBuf.SetData(m.Payload)

	err := w32.NtAlpcSendWaitReceivePort(
		m.Handle,
		m.Flags,
		msgBuf,
		pSendAttrs,
		msgBuf,
		nil,
		nil,
		nil,
	)

	// Even if there's an error we still fill the reply, in case there was
	// some kind of respose
	reply.PORT_MESSAGE = msgBuf.PORT_MESSAGE
	reply.Payload = msgBuf.GetData()
	if err != nil {
		log.Printf("error %v in Send", err)
		return err
	}
	return nil

}

func main() {

	alpc := new(ALPC)

	server := rpc.NewServer()
	server.Register(alpc)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Printf("listening on %s", listener.Addr().String())

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established: %s\n", conn.RemoteAddr().String())
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}

}
