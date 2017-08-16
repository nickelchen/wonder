package alice

import (
	"encoding/json"
	"fmt"
	"github.com/nickelchen/wonder/land"
	"github.com/nickelchen/wonder/share"

	log "github.com/sirupsen/logrus"
)

type InfoResponseStream struct {
	client *IPCClient
	seq    uint64
}

func newInfoResponseStream(client *IPCClient, seq uint64) *InfoResponseStream {
	s := InfoResponseStream{
		client: client,
		seq:    seq,
	}

	return &s
}

func (s *InfoResponseStream) stream(infoResult *land.InfoResult) {
	resultCh := infoResult.ResultCh()
	respHeader := share.ResponseHeader{
		Seq:   s.seq,
		Error: "",
	}

	for {
		select {
		case obj := <-resultCh:
			bs, err := json.Marshal(obj.Item)

			if err != nil {
				log.Error(fmt.Sprintf("can not convert struct to bytes: %s", err))
				break
			}

			respBody := share.InfoResponseObj{
				Type:    obj.Type,
				Payload: bs,
			}
			if err := s.client.send(&respHeader, &respBody); err != nil {
				return
			}
		}
	}

}
