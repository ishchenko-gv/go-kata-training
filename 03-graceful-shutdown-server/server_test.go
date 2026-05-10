package server_test

import (
	server "graceful-shutdown-server"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	msgCh := make(chan string)
	workerPool := server.NewWorkerPool(msgCh, 10)
	broker := server.NewBroker(msgCh)
	s := server.NewServer(workerPool, broker, server.WithServerPort(3001))
	var readyCh = make(chan struct{})
	go func() {
		s.Start()
		readyCh <- struct{}{}
	}()
	<-readyCh

	req, err := http.NewRequest(http.MethodPost, "http://localhost:3001/schedule-job", nil)
	assert.NoError(t, err)

	c := http.DefaultClient
	c.Do(req)
}
