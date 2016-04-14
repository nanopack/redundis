package redundis

import (
	"fmt"
	"net"
	"time"

	"github.com/nanopack/redundis/config"
)

// dial continuously tries connecting to a remote for 10 seconds
func dial(srv *net.Conn) (net.Conn, error) {
	userGone := make(chan bool)

	// fetch address of master
	addr, err := getMaster()
	if err != nil {
		return nil, fmt.Errorf("Failed to get address of master - %v", err)
	}

	config.Log.Debug("Dialing '%v'...", addr)
	timeout := time.After(10 * time.Second)

	// loop attempts for timeout, allows dead endpoints to start back up
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("Timed out trying to connect to '%v'", addr)
		case <-userGone:
			return nil, fmt.Errorf("User disconnected before endpoint reached")
		default:
			if *srv == nil {
				config.Log.Debug("Client gone...")
				userGone <- true
				continue
			}

			cli, err := net.DialTimeout("tcp", addr, 5*time.Second)
			if err != nil {
				config.Log.Debug("Failed to connect to '%v' - %v", addr, err)
				time.Sleep(time.Second)
				continue
			}

			return cli, nil
		}
	}
}
