package redundis_test

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"

	"github.com/nanopack/redundis/config"
	"github.com/nanopack/redundis/core"
)

func TestMain(m *testing.M) {
	// manually configure
	initialize()

	rtn := m.Run()

	os.Exit(rtn)
}

func TestUseRedundis(t *testing.T) {
	// dial running redundis
	r, err := redis.DialURL("redis://127.0.0.1:6379", redis.DialConnectTimeout(config.TimeoutNotReady))
	if err != nil {
		t.Errorf("Failed to reach redis at: '127.0.0.1:6379'")
	}

	out, err := redis.Bytes(r.Do("INFO", "replication"))
	if err != nil {
		t.Errorf("Failed to get INFO - %v", err.Error())
	}

	if string(out) != "role:master" {
		t.Errorf("Unexpected output: %q", string(out))
	}
}

func initialize() {
	fmt.Println("Starting fake redis and sentinel...")
	go fakeSentinel(":26379")
	go fakeRedis(":6380")

	fmt.Println("Starting redundis...")
	go redundis.Start()
	time.Sleep(time.Second)
}

func fakeRedis(address string) error {
	serverSocket, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	for {
		conn, err := serverSocket.Accept()
		if err != nil {
			return err
		}
		go sayIt(&conn, "$11\r\nrole:master\r\n")
	}
	return nil
}

func fakeSentinel(address string) error {
	serverSocket, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	for {
		conn, err := serverSocket.Accept()
		if err != nil {
			return err
		}
		go sayIt(&conn, "*2\r\n$9\r\n127.0.0.1\r\n$4\r\n6380\r\n")
	}
	return nil
}

func sayIt(conn *net.Conn, reply string) {
	r := bufio.NewReader(*conn)

	for {
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return
		}

		if line == "\n" {
			continue
		}

		// hi, my name is tom
		(*conn).Write([]byte(reply))
	}
}
