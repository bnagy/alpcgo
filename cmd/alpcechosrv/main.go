package main

import (
	"github.com/bnagy/alpcgo"
	"github.com/bnagy/w32"
	"log"
	"time"
	"unsafe"
)

const (
	srvName = "\\Sessions\\1\\BaseNamedObjects\\GoEchoSrv"
)

func main() {

	// Set up server port
	log.Printf("trying to create port %s", srvName)
	hServerConn, err := alpcgo.CreatePort(srvName)
	if err != nil {
		log.Fatalf("unable to create server: %v", err)
	}
	log.Printf("server port created with handle %x", hServerConn)

	// Reusable buffers
	context := &w32.ALPC_CONTEXT_ATTR{}
	lump := [256]byte{} // random number that's bigger than an ALPC_CONTEXT_ATTR + attributes header
	recvMsg := w32.NewAlpcShortMessage()
	sendMsg := w32.NewAlpcShortMessage()
	pRecvAttrs := (*w32.ALPC_MESSAGE_ATTRIBUTES)(unsafe.Pointer(&lump[0]))
	pRecvAttrs.AllocatedAttributes = w32.ALPC_MESSAGE_CONTEXT_ATTRIBUTE
	pRecvAttrs.ValidAttributes = w32.ALPC_MESSAGE_CONTEXT_ATTRIBUTE

	// Keeping refs in this map "tricks" the GC into not reaping the
	// pointers to the port contexts we create in the receive loop
	handles := make(map[*w32.AlpcPortContext]struct{})

	for {

		recvMsg.Reset() // resets the TotalLength so we don't get buffer too small errors

		// All messages arrive on the Server Connection port
		_, err := alpcgo.Recv(hServerConn, &recvMsg, pRecvAttrs, nil)
		if err != nil {
			log.Fatalf("recv: error: %v", err)
		}

		if recvMsg.Type&w32.LPC_CONNECTION_REQUEST == w32.LPC_CONNECTION_REQUEST {

			log.Printf("connection Message: % x", recvMsg.GetData())

			portContext := w32.AlpcPortContext{}
			handles[&portContext] = struct{}{}
			hServerComm, err := alpcgo.Accept(hServerConn, &portContext, &recvMsg, true)
			if err != nil {
				log.Fatalf("failed to accept client: %v", err)
			}
			// Save the communication port handle in the context. We could
			// save anything we wanted, this is an opaque blob.
			portContext.Handle = hServerComm
			log.Printf("new Communication Port, handle: %x", hServerComm)

		} else {

			log.Printf("message: Type: %x Data: %s Continuation: %v",
				recvMsg.Type,
				string(recvMsg.GetData()),
				recvMsg.Type&w32.LPC_CONTINUATION_REQUIRED > 0,
			)

			pMsgAttrs := w32.AlpcGetMessageAttribute(
				pRecvAttrs,
				w32.ALPC_MESSAGE_CONTEXT_ATTRIBUTE,
			)

			if pMsgAttrs != nil {

				context = (*w32.ALPC_CONTEXT_ATTR)(pMsgAttrs)
				commHandle := context.PortContext.Handle

				if commHandle != 0 {

					if recvMsg.Type == w32.LPC_PORT_CLOSED || recvMsg.Type == w32.LPC_CLIENT_DIED {
						log.Printf("client on handle 0x%x is gone.", commHandle)

						e := w32.NtAlpcDisconnectPort(commHandle, 0)
						res := w32.CloseHandle(commHandle) // kernel32!CloseHandle
						if e == nil && res == true {
							log.Printf("communication port 0x%x cleaned up and closed.", commHandle)
						} else {
							log.Printf("error cleaning up. kernel32!CloseHandle: %v NtAlpcDisconnectPort: %v", e, res)
						}

						// Clean up any context resources here
						if _, found := handles[context.PortContext]; !found {
							log.Printf("couldn't find 0x%x in handle map?", context.PortContext)
						}
						delete(handles, context.PortContext)

						continue
					}

					msg := []byte(time.Now().String())
					// If we don't reset the PORT_MESSAGE header fields then
					// we can get all kinds of weird failures when we try to
					// set assorted flags and things
					sendMsg.Reset()
					sendMsg.SetData(msg)
					log.Printf("sending response \"%s\" to handle 0x%x", string(msg), commHandle)

					e := alpcgo.Send(
						commHandle,
						&sendMsg,
						w32.ALPC_MSGFLG_RELEASE_MESSAGE, // Send response as an LPC_DATAGRAM
						pRecvAttrs,
						nil,
					)

					if e != nil {
						log.Printf("failed to respond: %v", e)
						continue
					}
				}

			} else {
				log.Fatalf("context was nil :%#v", pMsgAttrs)
			}

		}

	}

}
