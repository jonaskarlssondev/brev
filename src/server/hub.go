package brev

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type hub struct {
	channels map[string]*channel
}

func newHub() *hub {
	h := &hub{
		channels: make(map[string]*channel),
	}

	return h
}

type registrationReq struct {
	// The name and identifier of the channel to register
	Channel string
}

func register(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.Header().Add("Accept", "application/json")
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can not parse request body.", http.StatusBadRequest)
			return
		}

		var req registrationReq
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Can not parse request body as valid request.", http.StatusBadRequest)
			return
		}

		if _, ok := s.hub.channels[req.Channel]; ok {
			http.Error(w, "Channel already exists.", http.StatusConflict)
			return
		}

		ch := newChannel(req.Channel)

		go s.hub.registerChannel(ch)

		s.router.HandleFunc(fmt.Sprintf("/publish/%s", req.Channel), publish(s.hub))
	}
}

func (h *hub) registerChannel(ch *channel) {
	h.channels[ch.name] = ch
}

type subscribeReq struct {
	Callback string
	Channel  string
}

func subscribe(h *hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.Header().Add("Accept", "application/json")
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Can not parse request body.", http.StatusBadRequest)
			return
		}

		var req subscribeReq
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Can not parse request body as valid request.", http.StatusBadRequest)
			return
		}

		ch, ok := h.channels[req.Channel]
		if !ok {
			http.Error(w, "Channel does not exist.", http.StatusNotFound)
			return
		}

		sub := subscriber{
			callback: req.Callback,
		}

		ch.subscribers = append(ch.subscribers, sub)
		h.channels[req.Channel] = ch
	}
}

func publish(h *hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Add("Allow", "POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/octet-stream" {
			w.Header().Add("Accept", "application/octet-stream")
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		params := strings.SplitAfter(r.URL.Path, "/publish/")
		if len(params) < 1 {
			http.Error(w, "No channel was specified.", http.StatusBadRequest)
			return
		}

		cName := params[1]

		ch, ok := h.channels[cName]
		if !ok {
			http.Error(w, fmt.Sprintf("Channel '%s' does not exist.", cName), http.StatusNotFound)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Could not read request body.", http.StatusBadRequest)
			return
		}

		m := &message{
			payload: payload,
		}

		ch.events.Append(*m)

		go h.persist(ch, m)

		go send(ch, m)
	}
}

func (h *hub) persist(ch *channel, m *message) {

}

func send(ch *channel, m *message) {
	for _, s := range ch.subscribers {
		go func(s *subscriber) {
			body := bytes.NewReader(m.payload)

			resp, err := http.Post(s.callback, "application/octet-stream", body)
			if err != nil {
				fmt.Printf("Failed to post message on channel '%s' to callback '%s'", ch.name, s.callback)
				return
			}

			if resp.StatusCode != 200 {
				fmt.Printf("Subscriber with callback '%s' responded '%d'.", s.callback, resp.StatusCode)
				return
			}
		}(&s)
	}
}
