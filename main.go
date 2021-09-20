package main

import (
	"flag"
	"log"
	server "jumpcloud_password_hash/server"
)

func main() {

	port := flag.Int( "port", 8080, "Port to listen on" )
	flag.Parse()

	log.Printf( "Starting server on port %d!", *port )
	server.HandleRequests( *port )
}