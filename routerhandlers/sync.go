package routerhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	redisclient "github.com/mattiapavese/broadway/redisclient"
)

type SynchronizeClientBody struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

func SynchronizeClient(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed.", http.StatusMethodNotAllowed)
		return
	}

	rdb := redisclient.GetRedisClient("localhost", "6379", "")

	ctx := context.Background()

	var body SynchronizeClientBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	toTombstone, err := rdb.SMembers(ctx, "host:"+body.Ip+":"+body.Port).Result()
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("Redis Error while executing op SMEMBERS; %s", err.Error()),
			http.StatusInternalServerError)
		return
	}

	if len(toTombstone) > 0 {
		if _, err := rdb.SAdd(ctx, TOMBSTONED, toTombstone).Result(); err != nil {
			http.Error(
				w,
				fmt.Sprintf("Redis Error while executing op SADD; %s", err.Error()),
				http.StatusInternalServerError)
			return
		}
		fmt.Printf("Tombstoned %d connections.\n", len(toTombstone))
	}

	respBody := struct {
		Op      string `json:"op"`
		Status  string `json:"status"`
		Details struct {
			Ip   string `json:"ip"`
			Port string `json:"port"`
		}
	}{
		Op:     "Synchronized machine.",
		Status: "OK",
		Details: struct {
			Ip   string `json:"ip"`
			Port string `json:"port"`
		}{
			Ip:   body.Ip,
			Port: body.Port,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respBody)
}
