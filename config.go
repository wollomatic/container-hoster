package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	HOSTLIST_PREFIX = "\n#-----BEGIN CONTAINER-HOSTER GENERATED CONTENT-----\n"
	HOSTLIST_INFO   = "# Please do not manually change this file while Container-Hoster ist running.\n#\n"
	HOSTLIST_SUFFIX = "#\n#-----END CONTAINER-HOSTER GENERATED CONTENT-----\n# Every content below this line will be deleted.\n"
	DOCKER_LABEL    = "de.wollomatic.container-hoster"
)

type config struct {
	refreshHostsfileInterval  time.Duration // Interval to check if the hosts file needs to be refreshed
	hostsfile                 string        // Path to the hosts file
	hostnameFromContainername bool          // if true, the container name will be used as hostname
	hostnameFromLabel         bool          // if true, the hostname will be taken from the label as defined in DOCKER_LABEL
	onlyLabeledContainers     bool          // if true, only containers with the label as defined in DOCKER_LABEL will be added to the hosts file
	logEvents                 bool          // if true, log docker events which cause a refresh of the hosts file
	networkRegexp             string        // if set, only containers with a network matching this regexp will be added to the hosts file
}

var conf = config{
	refreshHostsfileInterval:  10 * time.Second,
	hostsfile:                 "/hosts",
	hostnameFromContainername: true,
	hostnameFromLabel:         false,
	onlyLabeledContainers:     false,
	logEvents:                 false,
	networkRegexp:             ".*",
}

// isTrue parses a string and into a boolean
func isTrue(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "y", "on", "enable", "enabled":
		return true, nil
	case "false", "0", "no", "n", "off", "disable", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", s)
	}
}

// getFromEnv checks if config environment variables are set and sets the config accordingly
func (c *config) getFromENV() {

	log.Printf("Checking environment variables for configuration")
	if EnvHostsfile, ok := os.LookupEnv("CH_HOSTSFILE"); ok {
		log.Printf("   found CH_HOSTSFILE=%s\n", EnvHostsfile)
		c.hostsfile = EnvHostsfile
	}

	if EnvInterval, ok := os.LookupEnv("CH_INTERVAL"); ok {
		log.Printf("   found CH_INTERVAL=%s\n", EnvInterval)
		EnvIntervalParsed, err := time.ParseDuration(EnvInterval)
		if err != nil {
			log.Printf("Error parsing CH_INTERVAL: %v. Using default value.\n", err)
		} else {
			c.refreshHostsfileInterval = EnvIntervalParsed
		}
	}

	if EnvHostnameFromContainername, ok := os.LookupEnv("CH_HOSTNAME_FROM_CONTAINERNAME"); ok {
		log.Printf("   found CH_HOSTNAME_FROM_CONTAINERNAME=%s\n", EnvHostnameFromContainername)
		if EnvHostnameFromContainernameParsed, err := isTrue(EnvHostnameFromContainername); err != nil {
			log.Printf("Error parsing CH_HOSTNAME_FROM_CONTAINERNAME: %v. Using default value.\n", err)
		} else {
			c.hostnameFromContainername = EnvHostnameFromContainernameParsed
		}
	}

	if EnvHostnameFromLabel, ok := os.LookupEnv("CH_HOSTNAME_FROM_LABEL"); ok {
		log.Printf("   found CH_HOSTNAME_FROM_LABEL=%s\n", EnvHostnameFromLabel)
		if EnvHostnameFromLabelParsed, err := isTrue(EnvHostnameFromLabel); err != nil {
			log.Printf("Error parsing CH_HOSTNAME_FROM_LABEL: %v. Using default value.\n", err)
		} else {
			c.hostnameFromLabel = EnvHostnameFromLabelParsed
		}
	}

	if EnvOnlyLabeledContainers, ok := os.LookupEnv("CH_ONLY_LABELED_CONTAINERS"); ok {
		log.Printf("   found CH_ONLY_LABELED_CONTAINERS=%s\n", EnvOnlyLabeledContainers)
		if EnvOnlyLabeledContainersParsed, err := isTrue(EnvOnlyLabeledContainers); err != nil {
			log.Printf("Error parsing CH_ONLY_LABELED_CONTAINERS: %v. Using default value.\n", err)
		} else {
			c.onlyLabeledContainers = EnvOnlyLabeledContainersParsed
		}
	}

	if EnvLogEvents, ok := os.LookupEnv("CH_LOG_EVENTS"); ok {
		log.Printf("   found CH_LOG_EVENTS=%s\n", EnvLogEvents)
		if EnvLogEventsParsed, err := isTrue(EnvLogEvents); err != nil {
			log.Printf("Error parsing CH_LOG_EVENTS: %v. Using default value.\n", err)
		} else {
			c.logEvents = EnvLogEventsParsed
		}

	}

	if EnvNetworkRegexp, ok := os.LookupEnv("CH_NETWORK_REGEXP"); ok {
		log.Printf("   found CH_NETWORK_REGEXP=%s\n", EnvNetworkRegexp)
		c.networkRegexp = EnvNetworkRegexp
	}
}

func (c *config)logConfig() {
	log.Printf("Configuration:")
	log.Printf("   hostsfile: %s", c.hostsfile)
	log.Printf("   refreshHostsfileInterval: %s", c.refreshHostsfileInterval)
	log.Printf("   hostnameFromContainername: %t", c.hostnameFromContainername)
	log.Printf("   hostnameFromLabel: %t", c.hostnameFromLabel)
	log.Printf("   onlyLabeledContainers: %t", c.onlyLabeledContainers)
	log.Printf("   logEvents: %t", c.logEvents)
	log.Printf("   networkRegexp: %s", c.networkRegexp)
}