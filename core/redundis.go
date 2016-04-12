// Package redundis contains the proxy logic as well as the sentinel client that
// keeps the user connected to the correct redis master node.
package redundis

import (
	"fmt"
	"net"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/redundis/config"
)

// Start begins the tcp listener for users to connect to.
func Start() error {
	serverSocket, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return err
	}

	config.Log.Info("Listening on '%v'", config.ListenAddress)

	// blocking listener
	for {
		conn, err := serverSocket.Accept()
		if err != nil {
			return err
		}
		config.Log.Debug("Got Connection")
		go handleConnection(&conn)
	}
	return nil
}

// getMaster gets the address of the master node from sentinel
func getMaster() (string, error) {
	config.Log.Debug("Contacting sentinel for address of master...")

	// connect to sentinel in order to query for the master address
	r, err := redis.DialURL("redis://"+config.SentinelAddress, redis.DialConnectTimeout(30*time.Second), redis.DialWriteTimeout(10*time.Second), redis.DialPassword(config.SentinelPassword))
	if err != nil {
		return "", fmt.Errorf("Failed to reach redis - %v", err)
	}

	// retrieve the master redis address
	addr, err := redis.Strings(r.Do("SENTINEL", "get-master-addr-by-name", config.MonitorName))
	if err != nil {
		return "", fmt.Errorf("Failed to get-master-addr-by-name - %v", err)
	}

	// cleanup after ourselves
	r.Close()

	// construct a useable address from sentinel's response
	masterAddr := fmt.Sprintf("%v:%v", addr[0], addr[1])
	config.Log.Debug("Master address: '%v'", masterAddr)

	// wait for redis to transition to master
	if err = verifyMaster(masterAddr, config.SentinelPassword); err != nil {
		return "", fmt.Errorf("Could not verify master - %v", err)
	}

	return masterAddr, nil
}

// verifyMaster verifies that the decided master node has fully transitioned
func verifyMaster(addr, pass string) error {
	// connect to redis in order to verify its state
	r, err := redis.DialURL("redis://"+addr, redis.DialConnectTimeout(30*time.Second), redis.DialPassword(pass))
	if err != nil {
		return fmt.Errorf("Failed to reach redis at: '%v'", addr)
	}

	// give redis some time to transition
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for redis to transition to master")
		default:
			// retrieve the master redis status (use `INFO replication` and parse if need to support redis <2.8.12)
			info, err := redis.Values(r.Do("ROLE"))
			if err != nil {
				return fmt.Errorf("Failed to get INFO - %v", err)
			}

			// check if node is master
			if len(info) > 0 && string(info[0].([]byte)) == "master" {
				return nil
			}
		}
	}

	// cleanup after ourselves
	r.Close()

	return nil
}
