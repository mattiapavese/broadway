package main

import (
	"fmt"
	"net/http"

	background "github.com/mattiapavese/broadway/background"
	routerhandlers "github.com/mattiapavese/broadway/routerhandlers"
)

func main() {
	// Set up routes

	go background.PurgeTombstoned()

	http.HandleFunc("/client/sync", routerhandlers.SynchronizeClient)

	http.HandleFunc("/conn/register", routerhandlers.RegisterConnection)
	http.HandleFunc("/conn/tombstone", routerhandlers.TombstoneConnection)

	http.HandleFunc("/chann/subscribe", routerhandlers.SubscribeConnOnChannel)
	http.HandleFunc("/chann/broadcast/json", routerhandlers.BroadcastJSONOnChannel)
	http.HandleFunc("/chann/subscribe-many", routerhandlers.SubscribeConnOnManyChannels)

	// Start the server
	fmt.Println("Broadway listening on port 8080.")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
