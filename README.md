# container-hoster
A simple "etc/hosts" file injection tool to resolve names of local Docker containers on the host. 


## Configuration
Since container-hoster is intended to be used in a Docker container, it is configured via environment variables. The following variables are available:

``CH_HOSTSFILE``: The path to the hosts file to be injected. Defaults to ``/hosts``. The real hostfile should be mounted as a bind mount to this path.

``CH_INTERVAL``: The interval in seconds to check if an update for the hostsfile is needed. It is formatted as a Go duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". Defaults to ``10s``.

``CH_HOSTNAME_FROM_CONTAINERNAME``: If set to true, the container name will be used as hostname. Defaults to ``true``.

``CH_HOSTNAME_FROM_LABEL``: If set to true, the value given in container label ``de.wollomatic.container-hoster.name`` will be used as hostname. Defaults to ``false``.

``CH_ONLY_LABELED_CONTAINERS``: If set to true, only containers with the label ``de.wollomatic.container-hoster.enable=true`` will be added to the hosts file and all other containers are ignored. Defaults to ``false``, so every container is added to the hosts file.

``CH_NETWORK_REGEXP``: A regular expression to match the network name of the container. Only containers with a matching network name will be added to the hosts file. Defaults to ``.*``.

``CH_LOG_EVENTS``: It set to true, all docker events which lead to rewrite the hosts file will be logged to stdout. Defaults to ``false``

## License

This project is licensed under the [MIT license](LICENSE.md)
Creative Commons License - see the [LICENSE](LICENSE) file for details

## Acknowledgments
Thanks to [David Darias](https://github.com/dvddarias) for the original idea [docker-hoster](https://github.com/dvddarias/docker-hoster).