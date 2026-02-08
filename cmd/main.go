package main

import (
	"IM/internal/controller"
	"IM/internal/service"
	"IM/internal/storage/cache"
	"log"
	"net/http"
)

func main() {
	cache.InitRedis()
	hub := service.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		controller.ServeWs(hub, w, r)
	})

	log.Println("IM Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
