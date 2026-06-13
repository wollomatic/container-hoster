package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// refreshHostsfile reads the docker containers, creates the container list, and writes the hosts file.
func refreshHostsfile(cli *client.Client) error {

	var engineHosts []byte

	result, err := cli.ContainerList(context.Background(), client.ContainerListOptions{})
	if err != nil {
		return err
	}

	engineHosts = append(engineHosts, []byte(HOSTLIST_PREFIX)...)
	engineHosts = append(engineHosts, []byte(HOSTLIST_INFO)...)

	if len(result.Items) > 0 {
		for _, c := range result.Items {
			if conf.onlyLabeledContainers && (strings.ToLower(c.Labels[CONTAINER_LABEL_PREFIX+".enabled"]) != "true") {
				// log.Println("Skipping c", c.Names[len(c.Names)-1], "because it is not labeled with", CONTAINER_LABEL_PREFIX+".enabled=true")
				continue
			}
			if strings.ToLower(c.Labels[CONTAINER_LABEL_PREFIX+".exclude"]) == "true" {
				// log.Println("Skipping c", c.Names[len(c.Names)-1], "because it is labeled with", CONTAINER_LABEL_PREFIX+".exclude=true")
				continue
			}
			containerHostList := getContainerHostList(c)
			if containerHostList != "" && c.NetworkSettings != nil {
				for networkName, networkInfo := range c.NetworkSettings.Networks {
					if networkRegexpCompiled.MatchString(networkName) && networkInfo.IPAddress.IsValid() {
						engineHosts = append(engineHosts, []byte(fmt.Sprintf("%-15s %-60s # %s\n", networkInfo.IPAddress.String(), containerHostList, networkName))...)
					}
				}
			}
		}
	}
	engineHosts = append(engineHosts, []byte(HOSTLIST_SUFFIX)...)

	return writeHostsfile(engineHosts)
}

// writeHostsfile reads the hosts file until the HOSTLIST_PREFIX and appends the given byte slice
func writeHostsfile(bs []byte) error {

	hf, err := os.ReadFile(conf.hostsfile)
	if err != nil {
		return err
	}

	hfnew := bytes.Split(hf, []byte(HOSTLIST_PREFIX))[0]
	hfnew = append(hfnew, bs...)

	return os.WriteFile(conf.hostsfile, hfnew, 0644) // #nosec G306 -- hostsfile has to be writable
}

// getContainerHostList returns the list of hostnames for a given container
func getContainerHostList(container container.Summary) string {
	var s string

	if conf.hostnameFromContainername {
		s = strings.TrimPrefix(container.Names[len(container.Names)-1], "/") + "  "
	}

	if label, ok := container.Labels[CONTAINER_LABEL_PREFIX+".name"]; ok && conf.hostnameFromLabel {
		s = s + label
	}

	return strings.Trim(s, " ")
}
