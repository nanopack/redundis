package redundis

import (
	"fmt"
	"io"
	"net"

	"github.com/nanopack/redundis/config"
)

// pipe connects the reader to the writer and closes both instances (all io.Copy()s)
// on fail or connection end.
func pipe(writer, reader *net.Conn, label string) {
	io.Copy(*writer, *reader)
	config.Log.Debug("%v hung up", label)
	// probably redundant, we get here because a failure to read in the io.Copy()
	(*reader).Close()

	// end the other relevant io.Copy()
	if *writer != nil {
		(*writer).Close()
	}

	// set reader to nil so dial doesn't continue trying to reach a dead endpoint
	*reader = nil
}

// handleConnection creates and connects a non-buffered pipe to the already
// existing server connection using a `net.Conn` from net.Pipe(). It then dials
// the endpoint and finishes the user->endpoint piping by connecting the other
// `net.Conn` to the client.
//
// NOTE: `net.Pipe()` is only neccessary for detecting a user hangup. Upon user
// disconnect, we prevent a possible DOS by ending the `dial()` function's
// ability to redial.
//
//  go pipe(&cli, srv, "User")
//  pipe(srv, &cli, "Endpoint")
// would be sufficient (following dial()) if we weren't concerned by the
// extra connection attempts made by dial.
func handleConnection(srv *net.Conn) {
	// Create and connect a non-buffered pipe to the already existing connection
	s, c := net.Pipe()
	config.Log.Debug("Piping user input to server...")
	go pipe(&s, srv, "User")
	config.Log.Debug("Piping server input to user...")
	go pipe(srv, &s, "Server")

	// dial the endpoint
	cli, err := dial(srv)
	if err != nil {
		config.Log.Error("Failed to contact endpoint - %v", err)
		(*srv).Write([]byte(fmt.Sprintf("Failed to contact endpoint - %v", err)))
		(*srv).Close()
		return
	}

	// Connect the non-buffered pipe to the dialed client
	config.Log.Debug("Piping client input to endpoint...")
	go pipe(&cli, &c, "Client")
	config.Log.Debug("Piping endpoint input to client...")
	pipe(&c, &cli, "Endpoint")

	config.Log.Debug("Piping session done")
}
