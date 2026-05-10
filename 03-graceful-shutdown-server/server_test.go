package server_test

import (
	server "graceful-shutdown-server"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	s := server.NewServer()
	var readyCh = make(chan struct{})
	go func() {
		s.Start()
		readyCh <- struct{}{}
	}()
	<-readyCh

	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/schedule-job", nil)
	assert.NoError(t, err)

	c := http.DefaultClient
	c.Do(req)
}
