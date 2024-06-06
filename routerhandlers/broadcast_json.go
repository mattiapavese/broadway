package routerhandlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	redisclient "github.com/mattiapavese/broadway/redisclient"
)

type BroadcastOnChannelJSONBody struct {
	Channkey     string          `json:"channkey"`
	HTTPCallback string          `json:"http-callback"`
	Message      json.RawMessage `json:"message"`
}

func broadcast(machine string, connkey string, httpCallback string, data json.RawMessage, done chan<- struct{}) {

	fmt.Println("debug")
	fmt.Println(machine, connkey, httpCallback)

	// machine and connkey are saved with prefixes host: and conn: in redis
	// remove those to get - host of / connection key on - broadway client
	machine_ := strings.SplitN(machine, ":", 2)[1]
	connkey_ := strings.SplitN(connkey, ":", 2)[1]
	var locationCallback string
	switch httpCallback {
	case "":
		locationCallback = "/broadway/default/receive-json"
	default:
		locationCallback = httpCallback
		if locationCallback[0] != '/' {
			locationCallback = "/" + locationCallback
		}
	}

	fmt.Printf("⏳posting to http://%s%s with connkey %s\n", machine, locationCallback, connkey)

	jsonData := struct {
		Connkey string          `json:"connkey"`
		Data    json.RawMessage `json:"data"`
	}{
		Connkey: connkey_,
		Data:    data,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}

	resp, err := http.Post("http://"+machine_+locationCallback, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Fatalf("Error making POST request: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("✅ posted to http://%s%s with connkey %s\n", machine, locationCallback, connkey)

	done <- struct{}{}

}

func BroadcastJSONOnChannel(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	var body BroadcastOnChannelJSONBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")

	connkeys, err := rdb.SMembers(ctx, "chann:"+body.Channkey).Result()
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing op SMEMBERS; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	if len(connkeys) == 0 {
		http.Error(
			w,
			fmt.Sprintf("Unable to broadcast: No connection key found for channel key %s.", "chann:"+body.Channkey),
			http.StatusBadRequest)
		return
	}

	machines, err := rdb.MGet(ctx, connkeys...).Result()
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing op MGET; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	done := make(chan struct{}, len(machines))
	fmt.Println("DEBUG")
	for idx := range machines {
		fmt.Println(machines[idx].(string), connkeys[idx])
		go broadcast(machines[idx].(string), connkeys[idx], body.HTTPCallback, body.Message, done)
	}

	//***If we make a channel and wait for http request we can obtain sync acknowledge of message
	//for real-time acknowledge this should be on
	// for i := 0; i < len(machines); i++ {
	// 	<-done
	// 	print("flushed.\n")
	// }

	respBody := struct {
		Op      string `json:"op"`
		Status  string `json:"status"`
		Details struct {
			Channkey string `json:"channkey"`
		}
	}{
		Op:     "Broadcasted channel on message.",
		Status: "OK",
		Details: struct {
			Channkey string `json:"channkey"`
		}{
			Channkey: body.Channkey,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)
}
