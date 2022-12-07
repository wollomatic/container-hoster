package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	PROGRAM_NAME = "container-hoster"
	PROGRAM_URL  = "github.com/wollomatic/container-hoster"
	VERSION      = "0.0.3"
)

var (
	refreshHostsfileNeeded bool = true
	networkRegexpCompiled  *regexp.Regexp
)

func main() {

	log.Printf("--- Starting %s %s (%s, %s, %s) %s ---\n", PROGRAM_NAME, VERSION, runtime.GOOS, runtime.GOARCH, runtime.Version(), PROGRAM_URL)

	conf.getFromENV()
	conf.logConfig()
	networkRegexpCompiled = regexp.MustCompile(conf.networkRegexp)

	// check if hostsfile is writable
	if _, err := os.OpenFile(conf.hostsfile, os.O_WRONLY, 0644); err != nil {
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
	defer cli.Close()

	// check if docker is running
	_, err = cli.Ping(context.Background())
	if err != nil {
		log.Fatalf("Docker error: %v", err)
	}

	// create background job for refreshing the hosts file
	fch := make(chan error)
	go refreshHostsfileJob(fch, conf.hostsfile, cli)

	// create listener for docker events
	ch, ech := cli.Events(context.Background(), types.EventsOptions{})

	log.Println("everything seems okay, waiting for things to happen")
	// wait for things to happen
	for {
		select {
		case event := <-ch:
			switch event.Action {
			case "start", "stop", "die", "destroy":
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
		case signal := <-done: // graceful shutdown
			log.Println("Received stop signal: ", signal)
			gracefulShutdown(0)
		}
	}

}

// refreshHostsfileJob is a background job for refreshing the hosts file
func refreshHostsfileJob(ech chan<- error, hostsfile string, cli *client.Client) {
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
