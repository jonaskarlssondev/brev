package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	args := os.Args

	if len(args) > 1 {
		if args[1] == "register" {
			register(args)
		} else if args[1] == "publish" {
			publish(args)
		} else if args[1] == "subscribe" {
			subscribe(args)
		}
	}
}

func register(args []string) {
	if len(args) < 3 {
		fmt.Printf("Too few arguments to register channel.")
		return
	}

	ch := struct {
		Channel string
	}{
		Channel: args[2],
	}

	payload, err := json.Marshal(ch)
	if err != nil {
		fmt.Printf("Could not marshal register request. Err: %s\n", err.Error())
		return
	}

	resp, err := http.Post("http://localhost:80/register", "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Printf("Failed to register '%s'. Err: %s\n", args[2], err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got error response %d from server when registering '%s'.\n", resp.StatusCode, args[2])
		return
	}

	fmt.Printf("Registered '%s'.\n", args[2])
}

type message struct {
	Val string
}

func publish(args []string) {
	if len(args) < 4 {
		fmt.Printf("Too few arguments to publish to channel.")
		return
	}

	msg := &message{
		Val: args[3],
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(msg)

	resp, err := http.Post(fmt.Sprintf("http://localhost:80/publish/%s", args[2]), "application/octet-stream", &buf)
	if err != nil {
		fmt.Printf("Failed to publish '%s'. Err: %s\n", args[2], err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got error response %d from server when publishing. '%s'.\n", resp.StatusCode, args[2])
		return
	}
}

func subscribe(args []string) {
	if len(args) < 3 {
		fmt.Printf("Too few arguments to subscribe to channel.")
		return
	}

	callback := "http://localhost:8001/subscriber"

	ch := struct {
		Channel  string
		Callback string
	}{
		Channel:  args[2],
		Callback: callback,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(ch)

	resp, err := http.Post("http://localhost:80/subscribe", "application/json", &buf)
	if err != nil {
		fmt.Printf("Failed to subscribe '%s'. Err: %s\n", args[2], err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got error response %d from server when subscribing. '%s'.\n", resp.StatusCode, args[2])
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/subscriber", write)

	http.ListenAndServe(":8001", mux)
}

func write(w http.ResponseWriter, r *http.Request) {
	var msg message

	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Printf("Could not read received body.")
		return
	}

	fmt.Println(msg.Val)
}
