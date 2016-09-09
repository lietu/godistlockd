// The main executable that will parse settings and start up godistlockd
package main

import (
	"flag"
	"github.com/lietu/godistlock/server"
	"fmt"
)

var clientPort = flag.Int("clients", 10000, "Port to bind to for client connections")
var relayPort = flag.Int("relays", 20000, "Port to bind to for relay connections")
var testing = flag.Bool("testing", false, "Enable testing stuff")

func main() {
	flag.Parse()

	server := server.NewServer()
	// TODO: Configure
	server.Id = fmt.Sprintf("server-on-port-%d", *relayPort)
	server.Version = "1.0.0"
	server.Testing = *testing
	server.Run(*clientPort, *relayPort)
}