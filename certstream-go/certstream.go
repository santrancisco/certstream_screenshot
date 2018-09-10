package certstream

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/jsonq"
	"github.com/pkg/errors"
)

func NewQueryCatchPanic(v interface{}) *jsonq.JsonQuery {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[ERROR] Recovered from ", r)
		}
	}()
	jq := jsonq.NewQuery(v)
	return jq
}

func CertStreamEventStream(skipHeartbeats bool) (chan jsonq.JsonQuery, chan error) {
	outputStream := make(chan jsonq.JsonQuery)
	errStream := make(chan error)

	go func() {
	MAINLOOP:
		for {
			c, _, err := websocket.DefaultDialer.Dial("wss://certstream.calidog.io", nil)

			if err != nil {
				errStream <- errors.Wrap(err, "Error connecting to certstream! Sleeping a few seconds and reconnecting... ")
				time.Sleep(5 * time.Second)
				continue
			}

			defer c.Close()
			defer close(outputStream)

			for {
				var v interface{}
				err = c.ReadJSON(&v)
				if err != nil {
					errStream <- errors.Wrap(err, "Error decoding json frame!")
				}
				var jq *jsonq.JsonQuery
				jq = NewQueryCatchPanic(v)
				if jq == nil {
					continue MAINLOOP
				}

				res, err := jq.String("message_type")
				if err != nil {
					errStream <- errors.Wrap(err, "Error creating jq object!")
				}

				if skipHeartbeats && res == "heartbeat" {
					continue
				}

				outputStream <- *jq
			}
		}
	}()

	return outputStream, errStream
}
