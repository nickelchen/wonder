package client

import (
	"github.com/nickelchen/wonder/share"
)

func (c *RPCClient) Plant(what, color string, number int, respCh chan<- string) error {
	seq := c.getSeq()

	header := share.RequestHeader{
		Seq:     seq,
		Command: share.PlantCommand,
	}
	request := share.PlantRequest{
		What:   share.PlantType(what),
		Color:  color,
		Number: number,
	}

	c.dispatch[seq] = &plantHandler{
		client: c,
		seq:    seq,
		respCh: respCh,
	}

	return c.send(&header, &request)
}

func (c *RPCClient) Info(respCh chan<- share.InfoResponseObj) error {
	seq := c.getSeq()

	header := share.RequestHeader{
		Seq:     seq,
		Command: share.InfoCommand,
	}

	request := share.InfoRequest{}

	initCh := make(chan error, 1)
	c.dispatch[seq] = &infoHandler{
		client: c,
		seq:    seq,
		init:   false,
		initCh: initCh,
		respCh: respCh,
	}

	if err := c.send(&header, &request); err != nil {
		delete(c.dispatch, seq)
		return err
	}

	// wait for first response
	select {
	case err := <-initCh:
		return err
	}
}

func (c *RPCClient) Subscribe(respCh chan<- share.EventResponseObj) error {
	seq := c.getSeq()

	header := share.RequestHeader{
		Seq:     seq,
		Command: share.SubscribeCommand,
	}

	request := share.SubscribeRequest{}

	initCh := make(chan error, 1)
	c.dispatch[seq] = &eventHandler{
		client: c,
		seq:    seq,
		init:   false,
		initCh: initCh,
		respCh: respCh,
	}

	if err := c.send(&header, &request); err != nil {
		delete(c.dispatch, seq)
		return err
	}

	// wait for first response
	select {
	case err := <-initCh:
		return err
	}
}
