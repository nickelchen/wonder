package alice

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"github.com/nickelchen/wonder/land"
	"github.com/nickelchen/wonder/share"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ugorji/go/codec"
)

type IPCClient struct {
	from                 string
	conn                 net.Conn
	reader               *bufio.Reader
	writer               *bufio.Writer
	dec                  *codec.Decoder
	enc                  *codec.Encoder
	eventResponseStreams map[uint64]*eventResponseStream
}

// send share.ResponseHeader and responseBody to client.
func (c *IPCClient) send(header *share.ResponseHeader, obj interface{}) error {
	if err := c.enc.Encode(header); err != nil {
		log.Error(fmt.Sprintf("Error in encode header: %s", err))
		log.Error(getStack())
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

type AliceIPC struct {
	alice    *Alice
	listener net.Listener
	clients  map[string]*IPCClient
	stop     bool
}

func NewAliceIPC(alice *Alice, listener net.Listener) *AliceIPC {
	ipc := &AliceIPC{
		alice:    alice,
		listener: listener,
		clients:  make(map[string]*IPCClient),
	}

	go ipc.listen()

	return ipc
}

func (i *AliceIPC) Shutdown() {
	if i.stop {
		return
	}
	i.stop = true

	i.listener.Close()
	for _, c := range i.clients {
		c.conn.Close()
	}
}

func (i *AliceIPC) listen() {
	for {
		conn, err := i.listener.Accept()
		if err != nil {
			return
		}
		client := &IPCClient{
			from:                 conn.RemoteAddr().String(),
			conn:                 conn,
			reader:               bufio.NewReader(conn),
			writer:               bufio.NewWriter(conn),
			eventResponseStreams: make(map[uint64]*eventResponseStream),
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
func (i *AliceIPC) handleClient(client *IPCClient) {
	log.Debug(fmt.Sprintf("Get client. %v", client))

	var reqHeader share.RequestHeader
	for {
		err := client.dec.Decode(&reqHeader)
		if err != nil {
			if err != io.EOF && !strings.Contains(err.Error(), "wsarecv") {
				log.Error(fmt.Sprintf("can not decode requstHeader: %s", err))
				log.Error(getStack())
			}
			return
		}
		var respHeader *share.ResponseHeader
		var respBody interface{}

		command := reqHeader.Command
		log.Debug(fmt.Sprintf("reqHeader.Command: %v", command))

		switch command {
		case share.PlantCommand:
			respHeader, respBody = i.handlePlant(client, reqHeader.Seq)
		case share.InfoCommand:
			respHeader, respBody = i.handleInfo(client, reqHeader.Seq)
		case share.SubscribeCommand:
			respHeader, respBody = i.handleSubscribe(client, reqHeader.Seq)
		}

		log.Debug(fmt.Sprintf("respHeader is :%v", respHeader))
		log.Debug(fmt.Sprintf("respBody is :%v", respBody))

		if respHeader != nil {
			client.send(respHeader, respBody)
		}
	}
}

func (i *AliceIPC) handlePlant(client *IPCClient, seq uint64) (*share.ResponseHeader, *share.PlantResponse) {
	var req share.PlantRequest
	if err := client.dec.Decode(&req); err != nil {
		return nil, nil
	}

	plantParams := land.PlantParams{
		What:   req.What,
		Color:  req.Color,
		Number: req.Number,
	}

	plantResult, err := i.alice.Plant(&plantParams)

	respHeader := share.ResponseHeader{
		Seq:   seq,
		Error: errorToString(err),
	}

	respBody := share.PlantResponse{
		Succ: plantResult.Succ,
		Fail: plantResult.Fail,
	}

	return &respHeader, &respBody
}

func (i *AliceIPC) handleInfo(client *IPCClient, seq uint64) (*share.ResponseHeader, *share.InfoResponse) {
	log.Debug(fmt.Sprintf("handleInfo start"))
	var req share.InfoRequest
	if err := client.dec.Decode(&req); err != nil {
		log.Error("can not decode infoRequest")
		return nil, nil
	}

	infoParams := land.InfoParams{}

	infoResult, err := i.alice.Info(&infoParams)

	if err == nil {
		infoRespStream := newInfoResponseStream(client, seq)
		go infoRespStream.stream(infoResult)
	}

	respHeader := share.ResponseHeader{
		Seq:   seq,
		Error: errorToString(err),
	}

	return &respHeader, nil
}

func (i *AliceIPC) handleSubscribe(client *IPCClient, seq uint64) (*share.ResponseHeader, *share.SubscribeResponse) {
	var req share.SubscribeRequest
	if err := client.dec.Decode(&req); err != nil {
		log.Error("can not decode subscribeRequest")
		return nil, nil
	}

	respHeader := share.ResponseHeader{
		Seq:   seq,
		Error: "",
	}
	if _, ok := client.eventResponseStreams[seq]; ok {
		respHeader.Error = "stream with seq already exists"
	}

	s := newEventResponseStream(client, seq)
	client.eventResponseStreams[seq] = s

	// prevent race condition. make sure send this response first, then we can stream events.
	// TODO
	defer i.alice.Subscribe(s)

	return &respHeader, nil
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}
