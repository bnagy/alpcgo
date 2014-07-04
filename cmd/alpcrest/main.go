package main

import (
	"fmt"
	"github.com/bnagy/alpcgo"
	"github.com/bnagy/w32"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
	"sync"
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

// pool of reusable message buffers
var msgPool = sync.Pool{New: func() interface{} { m := w32.NewAlpcShortMessage(); return &m }}

// handles that appear to be currently unclosed
var handles = make(map[w32.HANDLE]struct{})

type ALPC struct{}

func (alpc *ALPC) Connect(r *http.Request, cr *ConnReq, reply *w32.HANDLE) (e error) {

	//log.Printf("new connection from %v", r.RemoteAddr)

	connMsg := msgPool.Get().(*w32.AlpcShortMessage)
	defer msgPool.Put(connMsg)
	connMsg.Reset()
	connMsg.SetData(cr.Msg)

	//log.Printf("connecting to %v with connection message of %v bytes", cr.Port, len(cr.Msg))

	hClientComm, e := alpcgo.ConnectPort(cr.Port, "", connMsg)
	if e != nil {
		//log.Printf("error %v in Connect", e)
		return
	}

	//log.Printf("client connected! Handle is 0x%X", hClientComm)

	handles[hClientComm] = struct{}{}
	*reply = hClientComm
	return

}

func (alpc *ALPC) Close(r *http.Request, h *w32.HANDLE, reply *bool) (e error) {

	if _, found := handles[*h]; found {

		delete(handles, *h)

		e = w32.NtAlpcDisconnectPort(*h, 0)
		if e != nil {
			//log.Printf("unable to disconnect port 0x%x: %v", *h, e)
			return
		}

		//log.Printf("disconnected, trying kernel32!CloseHandle... [%v]", w32.CloseHandle(*h))
		w32.CloseHandle(*h)
		*reply = true
		return
	}

	e = fmt.Errorf("invalid handle 0x%X", *h)
	return
}

func (alpc *ALPC) SendRecv(r *http.Request, m *Message, reply *Message) error {

	if _, found := handles[m.Handle]; !found {
		//log.Printf("invalid handle 0x%X in Send", m.Handle)
		return fmt.Errorf("invalid handle 0x%X", m.Handle)
	}

	//log.Printf("%v is sending %v bytes to handle 0x%X", r.RemoteAddr, len(m.Payload), m.Handle)

	msg := msgPool.Get().(*w32.AlpcShortMessage)
	defer msgPool.Put(msg)
	msg.PORT_MESSAGE = m.PORT_MESSAGE
	msg.SetData(m.Payload)

	var pSendAttrs *w32.ALPC_MESSAGE_ATTRIBUTES
	if len(m.Attributes) > 0 {
		pSendAttrs = (*w32.ALPC_MESSAGE_ATTRIBUTES)(unsafe.Pointer(&m.Attributes[0]))
	}

	err := w32.NtAlpcSendWaitReceivePort(
		m.Handle,
		m.Flags,
		msg,
		pSendAttrs,
		msg,
		nil,
		nil,
		nil,
	)

	// Even if there's an error we still fill the reply, in case there was
	// some kind of respose
	reply.PORT_MESSAGE = msg.PORT_MESSAGE
	reply.Payload = msg.GetData()
	if err != nil {
		//log.Printf("error %v in Send", err)
		return err
	}
	//log.Printf("forwarding reply, type: %x, payload: %v bytes", msg.Type, msg.DataLength)
	return nil

}

func (alpc *ALPC) Send(r *http.Request, m *Message, reply *bool) error {

	if _, found := handles[m.Handle]; !found {
		//log.Printf("invalid handle 0x%X in Send", m.Handle)
		return fmt.Errorf("invalid handle 0x%X", m.Handle)
	}

	//log.Printf("%v is sending %v bytes to handle 0x%X", r.RemoteAddr, len(m.Payload), m.Handle)

	msg := msgPool.Get().(*w32.AlpcShortMessage)
	defer msgPool.Put(msg)
	msg.PORT_MESSAGE = m.PORT_MESSAGE
	msg.SetData(m.Payload)

	var pSendAttrs *w32.ALPC_MESSAGE_ATTRIBUTES
	if len(m.Attributes) > 0 {
		pSendAttrs = (*w32.ALPC_MESSAGE_ATTRIBUTES)(unsafe.Pointer(&m.Attributes[0]))
	}

	err := w32.NtAlpcSendWaitReceivePort(
		m.Handle,
		m.Flags,
		msg,
		pSendAttrs,
		nil,
		nil,
		nil,
		nil,
	)

	if err != nil {
		//log.Printf("error %v in Send", err)
		return err
	}
	//log.Printf("send was successful...")
	*reply = true

	return nil

}

func main() {

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(ALPC), "")
	http.Handle("/rpc", s)
	log.Println("listening to jsonrpcv2 on 0.0.0.0:1234/rpc")
	log.Fatal(http.ListenAndServe(":1234", nil))
}
