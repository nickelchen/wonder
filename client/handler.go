package client

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"nickelchen/wonder/share"
)

type seqHandler interface {
	Handle(*share.ResponseHeader)
	Cleanup()
}

type plantHandler struct {
	client *RPCClient
	seq    uint64
	respCh chan<- string
}

func (h *plantHandler) Handle(respHeader *share.ResponseHeader) {
	var resp share.PlantResponse
	if err := h.client.dec.Decode(&resp); err != nil {
		fmt.Printf("Error in decode resp string: %s\n", err)
		return
	}
	var ret = fmt.Sprintf("succ: %d, fail: %d", resp.Succ, resp.Fail)
	log.Printf("Get ret: %s\n", ret)

	// write to respCh
	select {
	case h.respCh <- ret:
	default:
		log.Info("plantHandler Dropping response, respCh full.")
	}
}

func (h *plantHandler) Cleanup() {
}

type infoHandler struct {
	client *RPCClient
	seq    uint64
	init   bool
	initCh chan error
	respCh chan<- share.InfoResponseObj
}

// get a response
func (h *infoHandler) Handle(respHeader *share.ResponseHeader) {
	if !h.init {
		h.init = true
		h.initCh <- strToError(respHeader.Error)
		return
	}

	var resp share.InfoResponseObj
	if err := h.client.dec.Decode(&resp); err != nil {
		fmt.Printf("Error in decode resp string: %s\n", err)
		return
	}

	// log.Info(fmt.Sprintf("get resp obj in Handle: %v\n", resp))

	select {
	case h.respCh <- resp:
	default:
		log.Info("infoHandler Dropping response, respCh full.")
	}
}

func (h *infoHandler) Cleanup() {
}

type eventHandler struct {
	client *RPCClient
	seq    uint64
	init   bool
	initCh chan error
	respCh chan<- share.EventResponseObj
}

// get a response
func (h *eventHandler) Handle(respHeader *share.ResponseHeader) {
	if !h.init {
		h.init = true
		h.initCh <- strToError(respHeader.Error)
		return
	}

	var resp share.EventResponseObj
	if err := h.client.dec.Decode(&resp); err != nil {
		fmt.Printf("Error in decode resp string: %s\n", err)
		return
	}

	// log.Info(fmt.Sprintf("get resp obj in Handle: %v\n", resp))

	select {
	case h.respCh <- resp:
	default:
		log.Info("eventHandler Dropping response, respCh full.")
	}
}

func (h *eventHandler) Cleanup() {
}
