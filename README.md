# Calliope
Email query and visualization for observation and study of wild email monsters

![angry faced envelope with arms waving](images/email_monster.png)


## Setting up Go and cloning the repo

Install using homebrew:

```
brew install go
brew install dep # Dependency management
```

Go requires your project directory to be a subdirectory of `$GOPATH/src`.
Example using `direnv` for user `me` on a Mac:


```bash
cd ~/
mkdir go
cd go
mkdir src
cd src
git clone https://github.com/oaktown/calliope.git
cd calliope
echo "export GOPATH=/Users/me/go" > .envrc
direnv allow
dep ensure
```

## Install Elasticsearch using Docker

This is adapted from olivere's [elastic-with-docker repo](https://github.com/olivere/elastic-with-docker).

In the project directory, type:

```bash
docker-compose -f docker-compose.local.yml up # this will run in the foreground
```

This will download the docker image (if you don't have it already) and run it on port 9200.
To test on the terminal:

```bash
curl http://localhost:9200
```

You should get a response like:

```json
{
  "name" : "EpNbYZk",
  "cluster_name" : "docker-cluster",
  "cluster_uuid" : "Z0St0x6PTPyprNKMhWhNrg",
  "version" : {
    "number" : "6.4.0",
    "build_flavor" : "oss",
    "build_type" : "tar",
    "build_hash" : "595516e",
    "build_date" : "2018-08-17T23:18:47.308994Z",
    "build_snapshot" : false,
    "lucene_version" : "7.4.0",
    "minimum_wire_compatibility_version" : "5.6.0",
    "minimum_index_compatibility_version" : "5.0.0"
  },
  "tagline" : "You Know, for Search"
}
```

To stop it, type:

```bash
docker-compose -f docker-compose.local.yml down
```

If you want data to persist (from [elastic-with-docker repo](https://github.com/olivere/elastic-with-docker)):

> Make sure to create a ./data directory locally and uncomment the volumes section in Docker Compose file(s) if you want your data to be persistent.

Note: `data` has been added to .gitignore

There's a Chrome extension called [ElasticSearch Head](https://chrome.google.com/webstore/detail/elasticsearch-head/ffmkiejjmecolpfloofpjologoblkegm) that you might find useful.

# Setup a Google Cloud project

TBD

