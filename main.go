package main

import (
	"context"
	"github.com/docker/docker/api/types/events"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/docker/docker/client"
)

const (
	programName = "container-hoster"
	programURL  = "github.com/wollomatic/container-hoster"
)

var (
	refreshHostsfileNeeded = true
	networkRegexpCompiled  *regexp.Regexp
	version                = "develop" // will be set in Github Action
)

func main() {

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

	// create docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	defer func(cli *client.Client) {
		err := cli.Close()
		if err != nil {
			log.Println("Error closing docker client:", err)
		}
	}(cli)

	// check if docker is running
	_, err = cli.Ping(context.Background())
	if err != nil {
		log.Fatalf("Docker error: %v", err)
	}

	// create background job for refreshing the hosts file
	fch := make(chan error)
	go refreshHostsfileJob(fch, cli)

	// create listener for docker events
	ch, ech := cli.Events(context.Background(), events.ListOptions{})

	log.Println("everything seems okay, waiting for things to happen")
	// wait for things to happen
	for {
		select {
		case event := <-ch:
			switch event.Action {
			case "start", "stop", "die", "destroy", "rename":
				if conf.logEvents {
					log.Println(event.Action, event.Actor.Attributes["name"])
				}
				refreshHostsfileNeeded = true
			}
		case err := <-ech:
			log.Println("Error updating hostsfile:", err)
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
		if refreshHostsfileNeeded {
			refreshHostsfileNeeded = false
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
