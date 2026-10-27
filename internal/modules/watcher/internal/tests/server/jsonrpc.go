package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RPCMocker struct {
	Method      string
	ParamsCheck func(t *testing.T, params []any)
	Result      string
}

func NewJSONRPCServer(t *testing.T, mocks ...RPCMocker) *httptest.Server {
	t.Helper()

	type rpcReq struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  []any  `json:"params"`
		ID      any    `json:"id"`
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.NoError(t, r.Body.Close())

		var req rpcReq
		assert.NoError(t, json.Unmarshal(body, &req))
		assert.NotEmpty(t, req.Method)

		for _, mock := range mocks {
			if mock.Method == req.Method {
				if mock.ParamsCheck != nil {
					mock.ParamsCheck(t, req.Params)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.Copy(w, strings.NewReader(mock.Result))

				return
			}
		}

		t.Fatalf("unexpected method: %s", req.Method)
	}))
}
