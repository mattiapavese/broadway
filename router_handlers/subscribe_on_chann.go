package routerhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	redisclient "github.com/mattiapavese/broadway/redis_client"
)

type SubscribeConnOnChannelBody struct {
	Connkey  string `json:"connkey"`
	Channkey string `json:"channkey"`
}

func SubscribeConnOnChannel(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	var body SubscribeConnOnChannelBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")

	pipe := rdb.TxPipeline()
	pipe.SAdd(ctx, "chann:"+body.Channkey, "conn:"+body.Connkey)
	pipe.SAdd(ctx, "conn:"+body.Connkey+":link", "chann:"+body.Channkey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing transaction; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	respBody := struct {
		Op      string `json:"op"`
		Status  string `json:"status"`
		Details struct {
			Connkey  string `json:"connkey"`
			Channkey string `json:"channkey"`
		}
	}{
		Op:     "Subscribed connection to channel.",
		Status: "OK",
		Details: struct {
			Connkey  string `json:"connkey"`
			Channkey string `json:"channkey"`
		}{
			Connkey:  body.Connkey,
			Channkey: body.Channkey,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)

}

type SubscribeConnOnManyChannelsBody struct {
	Connkey   string   `json:"connkey"`
	Channkeys []string `json:"channkeys"`
}

func SubscribeConnOnManyChannels(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	var body SubscribeConnOnManyChannelsBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")
	ctx := context.Background()
	for idx := range body.Channkeys {
		body.Channkeys[idx] = "chann:" + body.Channkeys[idx]
	}
	pipe := rdb.TxPipeline()
	pipe.SAdd(ctx, "conn:"+body.Connkey+":link", body.Channkeys)
	for _, channkey := range body.Channkeys {
		pipe.SAdd(ctx, channkey, "conn:"+body.Connkey)
	}
	_, err := pipe.Exec(ctx)

	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing transaction; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	respBody := struct {
		Op      string `json:"op"`
		Status  string `json:"status"`
		Details struct {
			Connkey   string   `json:"connkey"`
			Channkeys []string `json:"channkeys"`
		}
	}{
		Op:     "Subscribed connection to multiple channels.",
		Status: "OK",
		Details: struct {
			Connkey   string   `json:"connkey"`
			Channkeys []string `json:"channkeys"`
		}{
			Connkey:   body.Connkey,
			Channkeys: body.Channkeys,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)

}
