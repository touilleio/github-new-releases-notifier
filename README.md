Github new releases notification
====

This tool watches configured repos on Github 
and notifies whenever a new release is published.

Tags can be filtered using regexp to focus only
on relevant releases.

The notification leverage [shoutrrr](https://github.com/containrrr/shoutrrr)
library and support all the channels shoutrrr supports, i.e. (non-exhaustively)
Slack, Telegram, Discord, email, etc...

# Building

```
make ensure
make build
make package
```

# Configuration

The following example polls github every 4 hours, and filters
on tag names for Golang and Kubernetes, on title for AdoptOpenJDK,
an no filter for Prometheus.
New releases are notified in a Slack channel.

```
projects:
  - projectUrl: "https://github.com/golang/go"
    tagFilter: "go\\d+(\\.\\d+){1,2}$"
  - projectUrl: "https://github.com/kubernetes/kubernetes"
    tagFilter: "v\\d+(\\.\\d+){2}$"
  - projectUrl: "https://github.com/prometheus/prometheus"
  - projectUrl: "https://github.com/AdoptOpenJDK/openjdk11-upstream-binaries"
    titleFilter: ".*GA Release.*"

pollFrequency: 4h

notification:
  uri: "slack://xxxx/yyyy/zzz"
```

# Docker compose example

The example is available [here](./docker-compose.yml).

```
version: "3"
services:
  gh-releases-notifier:
    image: "touilleio/github-new-releases-notifier:v1"
    restart: unless-stopped
    ports:
      - "8080"
    volumes:
      - ./gh-releases-notifier-config.yml:/gh-releases-notifier-config.yml
      - ./gh-releases-notifier-data:/data
    environment:
      - CONFIG_FILE_PATH=/gh-releases-notifier-config.yml
      - DB_STORAGE_PATH=/data/bolt.db
      - NOTIFY_ALL_TAGS=false
      - LOG_LEVEL=debug
```

## Environment variables

| Name | Default | Description |
|------|---------|-------------|
| `CONFIG_FILE_PATH` | | Path inside the container of the config file as shown above. Example /gh-releases-notifier-config.yml |
| `DB_STORAGE_PATH` | | Path of the boltDB file remembering which tag was already seen. Example: /data/bolt.db |
| `NOTIFY_ALL_TAGS` | false | For newly added Github repository, should all the tags being notified? |
| `LOG_LEVEL` | info | Logging verbosity level |
| `PORT` | 8080 | Port to listen to for the HTTP endpoints |

## Endpoints

*/list*

List all the tags already observed.

```
curl http://localhost:8080/list | jq
```

*/delete*

Delete a specific tag so that it will be (re-)notified

```
curl -X PUT -d '{"repo": "https://github.com/golang/go","tag": "go1.16"}' http://localhost:8080/delete
```

*/metrics*

Prometheus metrics exposure

*/debug/verbosity*

Special endpoint to change logging verbosity at runtime.

```
curl http://localhost:8080/debug/verbosity
# returns current verbosity
curl -X PUT -d debug http://localhost:8080/debug/verbosity
# change verbosity to debug
```

# Behind the scene

Github releases polling relies on [Atom]() feed provided by Github, 
for instance [https://github.com/kubernetes/kubernetes/releases.atom](https://github.com/kubernetes/kubernetes/releases.atom).

The atom feed got parsed using [Gofeed](https://github.com/mmcdole/gofeed) library, 
and [shoutrrr](https://github.com/containrrr/shoutrrr) for the dispatching.

Persistence layer is implemented using [BoltDB](https://github.com/etcd-io/bbolt).
