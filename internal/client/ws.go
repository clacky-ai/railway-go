package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/railwayapp/cli/internal/config"
)

// GraphQL WS messages per graphql-transport-ws
const (
	wsTypeConnectionInit = "connection_init"
	wsTypeConnectionAck  = "connection_ack"
	wsTypeSubscribe      = "subscribe"
	wsTypeNext           = "next"
	wsTypeError          = "error"
	wsTypeComplete       = "complete"
	wsTypePing           = "ping"
	wsTypePong           = "pong"
)

type wsMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type subscribePayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type nextPayload struct {
	Data   json.RawMessage   `json:"data"`
	Errors []json.RawMessage `json:"errors"`
}

// Subscribe opens a graphql-transport-ws subscription and yields raw data frames via callback until complete or ctx done.
func Subscribe(ctx context.Context, cfg *config.Config, query string, variables map[string]interface{}, onData func(data json.RawMessage), onError func(err error)) error {
	// Build URL
	u := url.URL{Scheme: "wss", Host: fmt.Sprintf("backboard.%s", cfg.GetHost()), Path: "/graphql/v2"}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 30 * time.Second,
		Subprotocols:     []string{"graphql-transport-ws"},
	}

	header := http.Header{}
	if t := config.GetRailwayToken(); t != nil {
		header.Set("project-access-token", *t)
	} else if t := cfg.GetRailwayAuthToken(); t != nil {
		header.Set("authorization", fmt.Sprintf("Bearer %s", *t))
	}
	// Helpful headers
	header.Set("x-source", fmt.Sprintf("railway-cli/%s", "4.6.1"))
	header.Set("user-agent", fmt.Sprintf("railway-cli/%s", "4.6.1"))

	conn, _, err := dialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return err
	}
	defer conn.Close()

	// connection_init
	initMsg := wsMessage{Type: wsTypeConnectionInit, Payload: json.RawMessage(`{}`)}
	if err := conn.WriteJSON(initMsg); err != nil {
		return err
	}

	// wait for connection_ack (and ignore pings)
	acked := false
	ackDeadline := time.NewTimer(10 * time.Second)
	for !acked {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ackDeadline.C:
			return errors.New("graphql ws: timeout waiting for connection_ack")
		default:
		}
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return err
		}
		switch msg.Type {
		case wsTypeConnectionAck:
			acked = true
		case wsTypePing:
			// reply pong
			_ = conn.WriteJSON(wsMessage{Type: wsTypePong, Payload: msg.Payload})
		case wsTypeError:
			return errors.New("graphql ws: connection error before ack")
		case wsTypeNext, wsTypeComplete:
			// ignore until ack
		}
	}

	// subscribe
	subID := fmt.Sprintf("%d", time.Now().UnixNano())
	subPayload, _ := json.Marshal(subscribePayload{Query: query, Variables: variables})
	subMsg := wsMessage{ID: subID, Type: wsTypeSubscribe, Payload: subPayload}
	if err := conn.WriteJSON(subMsg); err != nil {
		return err
	}

	// read loop
	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteJSON(wsMessage{ID: subID, Type: wsTypeComplete})
			return ctx.Err()
		default:
		}
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if onError != nil {
				onError(err)
			}
			return err
		}
		switch msg.Type {
		case wsTypePing:
			_ = conn.WriteJSON(wsMessage{Type: wsTypePong, Payload: msg.Payload})
		case wsTypeNext:
			var np nextPayload
			_ = json.Unmarshal(msg.Payload, &np)
			if len(np.Errors) > 0 {
				if onError != nil {
					onError(fmt.Errorf("subscription error"))
				}
				continue
			}
			if onData != nil {
				onData(np.Data)
			}
		case wsTypeError:
			if onError != nil {
				onError(fmt.Errorf("subscription error frame"))
			}
		case wsTypeComplete:
			return nil
		}
	}
}
