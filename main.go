package main

import (
	"log"
	server "jumpcloud_takehome/server"
)

func main() {
	log.Printf("Starting server on port 8080")
	server.HandleRequests(8080)
	log.Printf("Service has shutdown")
}