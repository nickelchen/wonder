package client

import (
	"bufio"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/nickelchen/wonder/share"

	log "github.com/sirupsen/logrus"

	"github.com/ugorji/go/codec"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func (c *RPCClient) getSeq() uint64 {
	return atomic.AddUint64(&c.seq, 1)
}

type Config struct {
	Addr    string
	Timeout time.Duration
}

type RPCClient struct {
	seq uint64

	timeout time.Duration
	conn    *net.TCPConn

	reader *bufio.Reader
	writer *bufio.Writer
	dec    *codec.Decoder
	enc    *codec.Encoder

	dispatch map[uint64]seqHandler
}

func ClientFromConfig(config *Config) (*RPCClient, error) {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	conn, err := net.DialTimeout("tcp", config.Addr, config.Timeout)
	if err != nil {
		return nil, err
	}
	client := RPCClient{
		seq:      1,
		conn:     conn.(*net.TCPConn),
		reader:   bufio.NewReader(conn),
		writer:   bufio.NewWriter(conn),
		timeout:  config.Timeout,
		dispatch: make(map[uint64]seqHandler),
	}

	client.dec = codec.NewDecoder(client.reader,
		&codec.MsgpackHandle{RawToString: true, WriteExt: true})
	client.enc = codec.NewEncoder(client.writer,
		&codec.MsgpackHandle{RawToString: true, WriteExt: true})

	go client.listen()
	return &client, nil
}

func (c *RPCClient) send(header *share.RequestHeader, obj interface{}) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}

	if err := c.enc.Encode(header); err != nil {
		return err
	}
	if obj != nil {
		if err := c.enc.Encode(obj); err != nil {
			return err
		}
	}
	if err := c.writer.Flush(); err != nil {
		return err
	}

	return nil
}

func (c *RPCClient) listen() {
	var respHeader share.ResponseHeader
	defer c.Close()
	for {
		if err := c.dec.Decode(&respHeader); err != nil {
			log.Error(err.Error())
			break
		}
		c.handleResponse(respHeader.Seq, &respHeader)
	}
}

func (c *RPCClient) handleResponse(seq uint64, respHeader *share.ResponseHeader) {
	handler, ok := c.dispatch[seq]
	if ok {
		handler.Handle(respHeader)
	}
}

func (c *RPCClient) Close() {
	c.conn.Close()
}

func strToError(s string) error {
	if s != "" {
		return fmt.Errorf(s)
	}
	return nil
}
