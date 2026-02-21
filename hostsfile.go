package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/docker/docker/client"
)

// refreshHostsfile reads the docker containers, creates the container list and starts writeHostsfile() to write the hosts file
func refreshHostsfile(cli *client.Client) error {

	var dockerHosts []byte

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return err
	}

	dockerHosts = append(dockerHosts, []byte(HOSTLIST_PREFIX)...)
	dockerHosts = append(dockerHosts, []byte(HOSTLIST_INFO)...)

	if len(containers) > 0 {
		for _, c := range containers {
			if conf.onlyLabeledContainers && (strings.ToLower(c.Labels[DOCKER_LABEL+".enabled"]) != "true") {
				// log.Println("Skipping c", c.Names[len(c.Names)-1], "because it is not labeled with", DOCKER_LABEL+".enabled=true")
				continue
			}
			if strings.ToLower(c.Labels[DOCKER_LABEL+".exclude"]) == "true" {
				// log.Println("Skipping c", c.Names[len(c.Names)-1], "because it is labeled with", DOCKER_LABEL+".exclude=true")
				continue
			}
			containerHostList := getContainerHostList(c)
			if containerHostList != "" {
				for networkName, networkInfo := range c.NetworkSettings.Networks {
					if networkRegexpCompiled.MatchString(networkName) && networkInfo.IPAddress != "" {
						dockerHosts = append(dockerHosts, []byte(fmt.Sprintf("%-15s %-60s # %s\n", networkInfo.IPAddress, containerHostList, networkName))...)
					}
				}
			}
		}
	}
	dockerHosts = append(dockerHosts, []byte(HOSTLIST_SUFFIX)...)

	return writeHostsfile(dockerHosts)
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

	if label, ok := container.Labels[DOCKER_LABEL+".name"]; ok && conf.hostnameFromLabel {
		s = s + label
	}

	return strings.Trim(s, " ")
}
