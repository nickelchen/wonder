package stage

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/nickelchen/wonder/share"

	log "github.com/sirupsen/logrus"

	"github.com/ugorji/go/codec"
)

type IPCClient struct {
	from   string
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	dec    *codec.Decoder
	enc    *codec.Encoder
}

// send share.ResponseHeader and responseBody to client.
func (c *IPCClient) send(header *share.ResponseHeader, obj interface{}) error {
	if err := c.enc.Encode(header); err != nil {
		log.Error(fmt.Sprintf("Error in encode header: %s", err))
		log.Error(trace())
		return err
	}

	if obj != nil {
		if err := c.enc.Encode(obj); err != nil {
			log.Error(fmt.Sprintf("Error in encode obj: %s", err))
			return err
		}
	}

	log.Info("in ipc.go send(), flushing")
	if err := c.writer.Flush(); err != nil {
		log.Error(fmt.Sprintf("Error in flush: %s", err))
		return err
	}

	return nil
}

type StageIPC struct {
	stage    *Stage
	listener net.Listener
	clients  map[string]*IPCClient
	stop     bool
}

func NewStageIPC(stage *Stage, listener net.Listener) *StageIPC {
	ipc := &StageIPC{
		stage:    stage,
		listener: listener,
		clients:  make(map[string]*IPCClient),
	}

	go ipc.listen()

	return ipc
}

func (i *StageIPC) Shutdown() {
	if i.stop {
		return
	}
	i.stop = true

	i.listener.Close()
	for _, c := range i.clients {
		c.conn.Close()
	}
}

func (i *StageIPC) listen() {
	for {
		conn, err := i.listener.Accept()
		if err != nil {
			return
		}
		client := &IPCClient{
			from:   conn.RemoteAddr().String(),
			conn:   conn,
			reader: bufio.NewReader(conn),
			writer: bufio.NewWriter(conn),
		}
		client.dec = codec.NewDecoder(client.reader,
			&codec.MsgpackHandle{RawToString: true, WriteExt: true})
		client.enc = codec.NewEncoder(client.writer,
			&codec.MsgpackHandle{RawToString: true, WriteExt: true})

		i.clients[client.from] = client

		go i.handleClient(client)
	}
}

// read client request header, dispatch command, send response to client.
func (i *StageIPC) handleClient(client *IPCClient) {
	log.Debug(fmt.Sprintf("Get client. %v", client))

	var reqHeader share.RequestHeader
	for {
		err := client.dec.Decode(&reqHeader)
		if err != nil {
			if err != io.EOF &&
				!strings.Contains(err.Error(), "wsarecv") &&
				!strings.Contains(err.Error(), "closed") {

				log.Error(fmt.Sprintf("can not decode requstHeader: %s", err))
				log.Error(trace())

			}
			return
		}
		var respHeader *share.ResponseHeader
		var respBody interface{}

		command := reqHeader.Command
		log.Debug(fmt.Sprintf("reqHeader.Command: %v", command))

		switch command {
		case share.ListServersCommand:
			respHeader, respBody = i.handleListServers(client, reqHeader.Seq)
		case share.ServerAliveCommand:
			respHeader, respBody = i.handleServerAlive(client, reqHeader.Seq)
		}

		log.Debug(fmt.Sprintf("respHeader is :%v", respHeader))
		log.Debug(fmt.Sprintf("respBody is :%v", respBody))

		if respHeader != nil {
			client.send(respHeader, respBody)
		}
	}
}

func (i *StageIPC) handleListServers(client *IPCClient, seq uint64) (*share.ResponseHeader, *share.ListServersResponse) {
	var req share.ListServersRequest
	if err := client.dec.Decode(&req); err != nil {
		return nil, nil
	}

	servers, err := i.stage.ListServers()
	respHeader := share.ResponseHeader{
		Seq:   seq,
		Error: errorToString(err),
	}

	respBody := share.ListServersResponse{
		Servers: servers,
	}

	return &respHeader, &respBody
}

func (i *StageIPC) handleServerAlive(client *IPCClient, seq uint64) (*share.ResponseHeader, *share.ServerAliveResponse) {
	var req share.ServerAliveRequest
	if err := client.dec.Decode(&req); err != nil {
		return nil, nil
	}

	msg, err := i.stage.ServerAlive(req.ServerAddr)

	respHeader := share.ResponseHeader{
		Seq:   seq,
		Error: errorToString(err),
	}

	respBody := share.ServerAliveResponse{
		Message: msg,
	}

	return &respHeader, &respBody
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
