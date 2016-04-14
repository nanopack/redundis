// Package redundis contains the proxy logic as well as the sentinel client that
// keeps the user connected to the correct redis master node.
package redundis

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/redundis/config"
)

var (
	// masterAddr is the cached address of the master node
	masterAddr = ""

	// discard is a dummy writer to aid in determining master node disconnections
	discard io.Writer = devNull(0)
)

// dummy writer type
type devNull int

// dummy write method
func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

// Start begins the tcp listener for users to connect to.
func Start() error {
	serverSocket, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return err
	}

	// keep an updated cache of the master node's address
	go watchMaster()

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

// getMaster returns the address of the cached master node
func getMaster() (string, error) {
	if masterAddr == "" {
		return "", fmt.Errorf("Address of master unknown")
	}
	return masterAddr, nil
}

// keep an updated cache of the master node's address
func watchMaster() {
	config.Log.Info("Monitoring master...")
	for {
		// update masterAddr
		err := updateMaster()
		if err != nil {
			config.Log.Error("Failed to update master - %v", err)
			time.Sleep(config.TimeoutSentinelPoll)
			continue
		}

		// connect to redis
		cli, err := net.DialTimeout("tcp", masterAddr, 5*time.Second)
		if err != nil {
			config.Log.Debug("Failed to connect to '%v' - %v", masterAddr, err)
			time.Sleep(config.TimeoutSentinelPoll)
			continue
		}

		// copy so we know when to check again for new master (the copy will break)
		config.Log.Debug("Watcher connected to master at '%v'", masterAddr)
		io.Copy(discard, cli)
		config.Log.Debug("Watcher disconnected")
	}
}

// updateMaster gets the address of the master node from sentinel
func updateMaster() error {
	config.Log.Debug("Contacting sentinel for address of master...")

	// connect to sentinel in order to query for the master address
	r, err := redis.DialURL("redis://"+config.SentinelAddress, redis.DialConnectTimeout(config.TimeoutNotReady), redis.DialReadTimeout(config.TimeoutSentinelPoll), redis.DialPassword(config.SentinelPassword))
	if err != nil {
		return fmt.Errorf("Failed to reach sentinel - %v", err)
	}

	// retrieve the master redis address
	addr, err := redis.Strings(r.Do("SENTINEL", "get-master-addr-by-name", config.MonitorName))
	if err != nil {
		return fmt.Errorf("Failed to get-master-addr-by-name - %v", err)
	}

	// cleanup after ourselves
	r.Close()

	// construct a useable address from sentinel's response
	masterAddr = fmt.Sprintf("%v:%v", addr[0], addr[1])
	config.Log.Debug("Master address: '%v'", masterAddr)

	// wait for redis to transition to master
	if err = verifyMaster(masterAddr, config.SentinelPassword); err != nil {
		return fmt.Errorf("Could not verify master - %v", err)
	}

	return nil
}

// verifyMaster verifies that the decided master node has fully transitioned
func verifyMaster(addr, pass string) error {
	// connect to redis in order to verify its state
	r, err := redis.DialURL("redis://"+addr, redis.DialConnectTimeout(config.TimeoutNotReady), redis.DialPassword(pass))
	if err != nil {
		return fmt.Errorf("Failed to reach redis at: '%v'", addr)
	}

	// give redis some time to transition
	timeout := time.After(config.TimeoutMasterWait)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Timed out waiting for redis to transition to master")
		default:
			// retrieve the redis node's role
			info, err := redis.Bytes(r.Do("INFO", "replication"))
			if err != nil {
				return fmt.Errorf("Failed to get INFO - %v", err)
			}

			// check if node is master
			if strings.Contains(string(info), "role:master") {
				return nil
			}
		}
	}

	// cleanup after ourselves
	r.Close()

	return nil
}
