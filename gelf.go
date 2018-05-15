package gelf

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/logspout/router"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
)

var hostname string

func init() {
	hostname, _ = os.Hostname()
	router.AdapterFactories.Register(NewGelfAdapter, "gelf")
}

// GelfAdapter is an adapter that streams UDP JSON to Graylog
type GelfAdapter struct {
	//	writer      *gelf.GelfWriter
	route       *router.Route
	adapterType string
}

// NewGelfAdapter creates a GelfAdapter with UDP as the default transport.
func NewGelfAdapter(route *router.Route) (router.LogAdapter, error) {

	// Identify adatper type to use later for building the write //
	routeAdapterType := route.AdapterType()
	_, found := router.AdapterTransports.Lookup(route.AdapterTransport(routeAdapterType))

	if !found {
		return nil, errors.New("unable to find adapter: " + route.Adapter)
	}

	// Will add checks here to identify what protocol is specified in adapter before using it //
	//gelfWriter, err := gelf.NewTCPWriter(route.Address)

	log.Println(route)

	return &GelfAdapter{
		route: route,
		//	writer:      gelfWriter,
		adapterType: routeAdapterType,
	}, nil
}

// Stream implements the router.LogAdapter interface.
func (a *GelfAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		m := &GelfMessage{message}
		level := gelf.LOG_INFO
		if m.Source == "stderr" {
			level = gelf.LOG_ERR
		}
		extra, err := m.getExtraFields()
		if err != nil {
			log.Println("Graylog:", err)
			continue
		}

		msg := gelf.Message{
			Version:  "1.1",
			Host:     hostname,
			Short:    m.Message.Data,
			TimeUnix: float64(m.Message.Time.UnixNano()/int64(time.Millisecond)) / 1000.0,
			Level:    level,
			RawExtra: extra,
		}
		// 	ContainerId:    m.Container.ID,
		// 	ContainerImage: m.Container.Config.Image,
		// 	ContainerName:  m.Container.Name,
		// }

		// here be message write.

		// Extra logic to identify what adapterType we have selected before writing
		// the message to graylog

		if a.adapterType == "tcp" {
			// Create and Stream using TCP Writer
			newWriter, err := gelf.NewTCPWriter(a.route.Address)
			//newWriter.GelfWriter = a.Writer
			checkError(err)
			sendMessage(newWriter, &msg)
		} else if a.adapterType == "udp" {
			// Create and Stream using the UDP Writer
			newWriter, err := gelf.NewUDPWriter(a.route.Address)
			//newWriter.GelfWriter = a.Writer
			checkError(err)
			sendMessage(newWriter, &msg)
		} else {
			// TLS is not supported so ignore message
			log.Println("Gelf Adapter: tls is not yet support")
		}

		/*if err := a.writer.WriteMessage(&msg); err != nil {
			log.Println("Graylog:", err)
			continue
		}*/
	}
}

func checkError(err error) {
	if err != nil {
		log.Println("Graylog:", err)
		continue
	}
}

// Wrapper implementing the interface //
func sendMessage(w gelf.Writer, m *gelf.Message) {
	if err := w.WriteMessage(m); err != nil {
		log.Println("Graylog:", err)
	}
}

// GelfMessage stores the router Message //
type GelfMessage struct {
	*router.Message
}

func (m GelfMessage) getExtraFields() (json.RawMessage, error) {

	extra := map[string]interface{}{
		"_container_id":   m.Container.ID,
		"_container_name": m.Container.Name[1:], // might be better to use strings.TrimLeft() to remove the first /
		"_image_id":       m.Container.Image,
		"_image_name":     m.Container.Config.Image,
		"_command":        strings.Join(m.Container.Config.Cmd[:], " "),
		"_created":        m.Container.Created,
	}
	for name, label := range m.Container.Config.Labels {
		if strings.ToLower(name[0:5]) == "gelf_" {
			extra[name[4:]] = label
		}
	}
	swarmnode := m.Container.Node
	if swarmnode != nil {
		extra["_swarm_node"] = swarmnode.Name
	}

	rawExtra, err := json.Marshal(extra)
	if err != nil {
		return nil, err
	}
	return rawExtra, nil
}
