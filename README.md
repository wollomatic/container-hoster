# container-hoster
A simple "etc/hosts" file injection tool to resolve names of local Docker containers on the host. It is inspired by [docker-hoster](https://github.com/dvddarias/docker-hoster) by [David Darias](https://github.com/dvddarias).


## Installation

The docker image is available on [Docker Hub](https://hub.docker.com/r/wollomatic/container-hoster/). A sample [docker-compose.yml](https://raw.githubusercontent.com/wollomatic/container-hoster/main/compose.yaml) file is provided in the [repository](https://github.com/wollomatic/container-hoster).

## Configuration
Container hoster is configured via environment variables. If no env variable is set, container-hoster will use the default value. The following variables are available:

* ``CH_HOSTSFILE``: The path to the hosts file to be injected. Defaults to ``/hosts``. The real hostsfile should be mounted as a bind mount to this path.

* ``CH_INTERVAL``: The interval in seconds to check if an update for the hostsfile is needed. It is formatted as a Go duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". Defaults to ``10s``.

* ``CH_HOSTNAME_FROM_CONTAINERNAME``: If set to true, the container name will be used as the hostname. Defaults to ``true``.

* ``CH_HOSTNAME_FROM_LABEL``: If set to true, the value given in container label ``de.wollomatic.container-hoster.name`` will be used as hostname. Defaults to ``false``.

* ``CH_ONLY_LABELED_CONTAINERS``: If set to true, only containers with the label ``de.wollomatic.container-hoster.enable=true`` will be added to the hosts file, and all other containers are ignored. Defaults to ``false``, so every container is added to the hosts file.

* ``CH_NETWORK_REGEXP``: A regular expression to match the network name of the container. Only containers with a matching network name will be added to the hosts file. Defaults to ``.*``.

* ``CH_LOG_EVENTS``: If set to true, all docker events which lead to rewrite the hosts file will be logged to stdout. Defaults to ``false``

## Container labels
Container labels are optional. The following labels are available:

* ``de.wollomatic.container-hoster.name``: The hostname to be used for the container if ``CH_HOSTNAME_FROM_LABEL`` is set to ``true``.

* ``de.wollomatic.container-hoster.enable``: If set to ``true``, the container will be added to the hosts file. Defaults to ``true``.

* ``de.wollomatic.container-hoster.exclude``: If set to ``true``, the container will be excluded from the hosts file.

## Security

In most cases, the container-hoster container will be run as root. This is necessary to be able to write to the hosts file and connect to the docker socket. Giving access to the docker socket is potentially dangerous because a container that has full access to the docker socket could start and stop containers. Container-hoster will only listen to docker events and will not start or stop containers. It will only update the hosts file if a container is started or stopped.

The build container image is made from scratch and contains no additional software. The dependencies are scanned with [trivy](https://github.com/aquasecurity/trivy-action).

Container-hoster does not need to have access to any network.

## License
This project is licensed under the [MIT license](LICENSE)
 - see the [LICENSE](LICENSE) file for details

## Acknowledgments
Thanks to [David Darias](https://github.com/dvddarias) for the original idea [docker-hoster](https://github.com/dvddarias/docker-hoster).