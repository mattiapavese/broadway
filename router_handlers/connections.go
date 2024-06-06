package routerhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	redisclient "github.com/mattiapavese/broadway/redis_client"
)

type RegisterConnectionBody struct {
	Ip      string `json:"ip"`
	Port    string `json:"port"`
	Connkey string `json:"connkey"`
}

func RegisterConnection(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	//validate the request body to match a predefined struct
	var body RegisterConnectionBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")

	var ctx = context.Background()

	//execute the transaction on redis db and handle eventual error
	host := body.Ip + ":" + string(body.Port)
	pipe := rdb.TxPipeline()
	pipe.Set(ctx, "conn:"+body.Connkey, "host:"+host, 0)
	pipe.SAdd(ctx, "host:"+host, "conn:"+body.Connkey)
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
			Connkey string `json:"connkey"`
			Ip      string `json:"ip"`
			Port    string `json:"port"`
		}
	}{
		Op:     "Registerd connection.",
		Status: "OK",
		Details: struct {
			Connkey string `json:"connkey"`
			Ip      string `json:"ip"`
			Port    string `json:"port"`
		}{
			Connkey: body.Connkey,
			Ip:      body.Ip,
			Port:    body.Port,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)
}

type TombstoneConnectionBody struct {
	Connkey string `json:"connkey"`
}

func TombstoneConnection(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")

	ctx := context.Background()

	var body TombstoneConnectionBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	_, err := rdb.SAdd(ctx, TOMBSTONED, "conn:"+body.Connkey).Result()

	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing op SADD; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	respBody := struct {
		Op      string `json:"op"`
		Status  string `json:"status"`
		Details struct {
			Connkey string `json:"connkey"`
		}
	}{
		Op:     "Tombstoned connection.",
		Status: "OK",
		Details: struct {
			Connkey string `json:"connkey"`
		}{
			Connkey: body.Connkey,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)

}
