package alice

import (
	"encoding/json"
	"fmt"
	"nickelchen/wonder/land"
	"nickelchen/wonder/share"

	log "github.com/sirupsen/logrus"
)

type eventResponseStream struct {
	//
	client  *IPCClient
	seq     uint64
	eventCh chan land.Event
}

func newEventResponseStream(client *IPCClient, seq uint64) *eventResponseStream {
	s := eventResponseStream{
		client:  client,
		seq:     seq,
		eventCh: make(chan land.Event, 512),
	}

	go s.stream()

	return &s
}

func (s *eventResponseStream) Handle(event land.Event) {
	// non-blocking send.
	select {
	case s.eventCh <- event:
	default:
		log.Warn(fmt.Sprintf("the stream event channel is full. dropping event: %s", event))
	}
}

func (s *eventResponseStream) stream() {
	respHeader := share.ResponseHeader{
		Seq:   s.seq,
		Error: "",
	}
	for {
		select {
		case event := <-s.eventCh:
			bs, err := json.Marshal(event.Item)
			if err != nil {
				log.Error(fmt.Sprintf("can not convert event item to bytes: %s", err))
			}

			respBody := share.EventResponseObj{
				Type:    event.Type,
				Payload: bs,
			}

			if err := s.client.send(&respHeader, &respBody); err != nil {
				return
			}
		}
	}
}
