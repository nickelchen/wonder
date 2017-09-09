package client

import (
	"github.com/nickelchen/wonder/share"
)

func (c *RPCClient) ServerAlive(serverAddr string, respCh chan<- string) error {
	seq := c.getSeq()

	header := share.RequestHeader{
		Seq:     seq,
		Command: share.ServerAliveCommand,
	}
	request := share.ServerAliveRequest{
		ServerAddr: serverAddr,
	}

	c.dispatch[seq] = &serverAliveHandler{
		client: c,
		seq:    seq,
		respCh: respCh,
	}

	return c.send(&header, &request)
}

func (c *RPCClient) ListServers(respCh chan<- []string) error {
	seq := c.getSeq()

	header := share.RequestHeader{
		Seq:     seq,
		Command: share.ListServersCommand,
	}
	request := share.ListServersRequest{}

	c.dispatch[seq] = &listServersHandler{
		client: c,
		seq:    seq,
		respCh: respCh,
	}

	return c.send(&header, &request)
}
