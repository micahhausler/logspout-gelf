package gelf

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/gliderlabs/logspout/router"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
	router.AdapterFactories.Register(NewGelfAdapter, "gelf")
}

// GelfAdapter is an adapter that streams UDP JSON to Graylog
type GelfAdapter struct {
	conn  net.Conn
	route *router.Route
}

// NewGelfAdapter creates a GelfAdapter with UDP as the default transport.
func NewGelfAdapter(route *router.Route) (router.LogAdapter, error) {
	transport, found := router.AdapterTransports.Lookup(route.AdapterTransport("udp"))
	if !found {
		return nil, errors.New("unable to find adapter: " + route.Adapter)
	}

	conn, err := transport.Dial(route.Address, route.Options)
	if err != nil {
		return nil, err
	}

	return &GelfAdapter{
		route: route,
		conn:  conn,
	}, nil
}

// Stream implements the router.LogAdapter interface.
func (a *GelfAdapter) Stream(logstream chan *router.Message) {
	for m := range logstream {

		msg := GelfMessage{
			Version:        "1.1",
			Host:           hostname,
			ShortMessage:   m.Data,
			Timestamp:      float64(m.Time.UnixNano()) / float64(time.Second),
			ContainerId:    m.Container.ID,
			ContainerImage: m.Container.Config.Image,
			ContainerName:  m.Container.Name,
		}
		js, err := json.Marshal(msg)
		if err != nil {
			log.Println("Graylog:", err)
			continue
		}
		_, err = a.conn.Write(js)
		if err != nil {
			log.Println("Graylog:", err)
			continue
		}
	}
}

type GelfMessage struct {
	Version      string  `json:"version"`
	Host         string  `json:"host"`
	ShortMessage string  `json:"short_message"`
	FullMessage  string  `json:"full_message,omitempty"`
	Timestamp    float64 `json:"timestamp,omitempty"`
	Level        int     `json:"level,omitempty"`

	ContainerId    string `json:"docker_container,omitempty"`
	ContainerImage string `json:"docker_image,omitempty"`
	ContainerName  string `json:"docker_name,omitempty"`
}
