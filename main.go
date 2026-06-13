package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/moby/moby/client"
)

const (
	programName = "container-hoster"
	programURL  = "github.com/wollomatic/container-hoster"
)

var (
	refreshHostsfileNeeded atomic.Bool
	networkRegexpCompiled  *regexp.Regexp
	version                = "develop" // will be set in Github Action
)

func main() {
	refreshHostsfileNeeded.Store(true)

	log.Printf("--- Starting %s %s (%s, %s, %s) %s ---\n", programName, version, runtime.GOOS, runtime.GOARCH, runtime.Version(), programURL)

	conf.getFromENV()
	conf.logConfig()
	networkRegexpCompiled = regexp.MustCompile(conf.networkRegexp)

	// check if hostsfile is writable
	_, err := os.OpenFile(conf.hostsfile, os.O_WRONLY, 0644) // #nosec G302 -- hostsfile needs 644 permissions
	if err != nil {
		log.Fatalf("Error: Hostsfile %s ist not writable: %s", conf.hostsfile, err)
	}

	// stop signal listener
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Create a client for the Moby Engine API.
	cli, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {
			log.Println("Error closing docker client:", err)
		}
	}(cli)

	// Check if the Moby API is reachable and negotiate a compatible API version.
	_, err = cli.Ping(context.Background(), client.PingOptions{NegotiateAPIVersion: true})
	if err != nil {
		log.Fatalf("Docker error: %v", err)
	}

	// create background job for refreshing the hosts file
	fch := make(chan error)
	go refreshHostsfileJob(fch, cli)

	// Create listener for docker events.
	eventResult := cli.Events(context.Background(), client.EventsListOptions{})

	log.Println("everything seems okay, waiting for things to happen")
	// wait for things to happen
	for {
		select {
		case event := <-eventResult.Messages:
			switch event.Action {
			case "start", "stop", "die", "destroy", "rename":
				if conf.logEvents {
					log.Println(event.Action, event.Actor.Attributes["name"])
				}
				refreshHostsfileNeeded.Store(true)
			}
		case err, ok := <-eventResult.Err:
			if !ok {
				log.Println("Docker event stream closed")
			} else {
				log.Println("Error updating hostsfile:", err)
			}
			gracefulShutdown(1)
		case err := <-fch:
			log.Println("Docker event Error:", err)
			gracefulShutdown(2)
		case sig := <-done: // graceful shutdown
			log.Println("Received stop signal: ", sig)
			gracefulShutdown(0)
		}
	}

}

// refreshHostsfileJob is a background job for refreshing the hosts file
func refreshHostsfileJob(ech chan<- error, cli *client.Client) {
	for {
		if refreshHostsfileNeeded.CompareAndSwap(true, false) {
			if conf.logEvents {
				log.Println("writing hosts file")
			}
			err := refreshHostsfile(cli)
			if err != nil {
				ech <- err
			}
		}
		time.Sleep(conf.refreshHostsfileInterval)
	}
}

// gracefulShutdown stops the program
func gracefulShutdown(i int) {
	err := writeHostsfile([]byte{})
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Println("exit with code", i)
	os.Exit(i)
}
