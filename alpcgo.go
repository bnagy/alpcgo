package alpcgo

import "github.com/bnagy/w32"

var basicPortAttr = w32.ALPC_PORT_ATTRIBUTES{
	MaxMessageLength: uint64(w32.SHORT_MESSAGE_MAX_SIZE),
	SecurityQos: w32.SECURITY_QUALITY_OF_SERVICE{
		Length:              w32.SECURITY_QOS_SIZE,
		ContextTrackingMode: w32.SECURITY_DYNAMIC_TRACKING,
		EffectiveOnly:       1,
		ImpersonationLevel:  w32.SecurityAnonymous,
	},
	Flags:          w32.ALPC_PORFLG_ALLOW_LPC_REQUESTS,
	DupObjectTypes: w32.ALPC_SYNC_OBJECT_TYPE,
}

func ObjectAttributes(name string) (oa w32.OBJECT_ATTRIBUTES, e error) {

	sd, e := w32.InitializeSecurityDescriptor(1)
	if e != nil {
		return
	}

	e = w32.SetSecurityDescriptorDacl(sd, nil)
	if e != nil {
		return
	}

	oa, e = w32.InitializeObjectAttributes(name, 0, 0, sd)
	return
}

func Send(
	hPort w32.HANDLE,
	msg *w32.AlpcShortMessage,
	flags uint32,
	pMsgAttrs *w32.ALPC_MESSAGE_ATTRIBUTES,
	timeout *int64,
) (e error) {

	e = w32.NtAlpcSendWaitReceivePort(hPort, flags, msg, pMsgAttrs, nil, nil, nil, timeout)
	return

}

func Recv(
	hPort w32.HANDLE,
	pMsg *w32.AlpcShortMessage,
	pMsgAttrs *w32.ALPC_MESSAGE_ATTRIBUTES,
	timeout *int64,
) (bufLen uint32, e error) {

	bufLen = uint32(pMsg.TotalLength)
	e = w32.NtAlpcSendWaitReceivePort(hPort, 0, nil, nil, pMsg, &bufLen, pMsgAttrs, timeout)
	return

}

// Convenience method to create an ALPC port with a NULL DACL. Requires an
// absolute port name ( where / is the root of the kernel object directory )
func CreatePort(name string) (hPort w32.HANDLE, e error) {

	oa, e := ObjectAttributes(name)
	if e != nil {
		return
	}

	hPort, e = w32.NtAlpcCreatePort(&oa, &basicPortAttr)

	return
}

func ConnectPort(serverName, clientName string, pConnMsg *w32.AlpcShortMessage) (hPort w32.HANDLE, e error) {

	oa, e := w32.InitializeObjectAttributes(clientName, 0, 0, nil)
	if e != nil {
		return
	}

	hPort, e = w32.NtAlpcConnectPort(
		serverName,
		&oa,
		&basicPortAttr,
		w32.ALPC_PORFLG_ALLOW_LPC_REQUESTS,
		nil,
		pConnMsg,
		nil,
		nil,
		nil,
		nil,
	)

	return
}

func Accept(
	hSrv w32.HANDLE,
	context *w32.AlpcPortContext,
	pConnReq *w32.AlpcShortMessage,
	accept bool,
) (hPort w32.HANDLE, e error) {

	oa, _ := w32.InitializeObjectAttributes("", 0, 0, nil)

	var accepted uintptr
	if accept {
		accepted++
	}

	hPort, e = w32.NtAlpcAcceptConnectPort(
		hSrv,
		0,
		&oa,
		&basicPortAttr,
		context,
		pConnReq,
		nil,
		accepted,
	)

	return
}
