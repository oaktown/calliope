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

We were inspired by https://github.com/vcollak/GmailContacts, which has nice
setup instructions.

# Build and run the app

When building the app `-i` will build all the dependencies and `-v` prints
out what its building.  The following commands will generate an executable
called `calliope` in the same directory and then run it:

```bash
go build -i -v
./calliope
```

Initial output:
```
====> Get ready to authenticate....

Open the link below in your browser. To give permission to view your email, click 'Allow' then copy the code...

https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=551888752777-k3ahicnth2t1n7c08jm2vhqempvi21ek.apps.googleusercontent.com&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fgmail.readonly&state=state-token

Paste the code here:
```

Paste the link into a browser and you will need to grant the app permission
to read your email.  Then it will display a new page with an access code.  If
you copy and paste into your terminal, it will print out some debugging logs.
(The app is a work in progress, and this is as far as we've gotten.)


```
Saving credential file to: oauth_token.json
```

Note: Next time, it will use the saved token instead of prompting you.

```
got client
2018/11/25 11:50:13 Retrieving messages starting on 2018/01/01
2018/11/25 11:50:14 Processing 100 messages...
```

It's actually truncating at 6 messages to allow for quicker iterations while
we figure out how to decode the messages and store in ElasticSearch

```
Sending Message ID: 1674c66d7eb92b56
2018/11/25 11:50:14 saving Message ID:  1674c66d7eb92b56
Sending Message ID: 1674c52e45369cb2
Sending Message ID: 1674c4fef3f29c46
Sending Message ID: 1674c43db8e5434b
2018/11/25 11:50:15 Indexed data id 1674c66d7eb92b56 to index mail, type document
2018/11/25 11:50:15 saving Message ID:  1674c52e45369cb2
Sending Message ID: 1674c3d0d75bb5a0
2018/11/25 11:50:15 Indexed data id 1674c52e45369cb2 to index mail, type document
2018/11/25 11:50:15 saving Message ID:  1674c4fef3f29c46
2018/11/25 11:50:15 Indexed data id 1674c4fef3f29c46 to index mail, type document
2018/11/25 11:50:15 saving Message ID:  1674c43db8e5434b
Sending Message ID: 1674c3c1f389215b
2018/11/25 11:50:15 Indexed data id 1674c43db8e5434b to index mail, type document
2018/11/25 11:50:15 saving Message ID:  1674c3d0d75bb5a0
2018/11/25 11:50:15 Indexed data id 1674c3d0d75bb5a0 to index mail, type document
2018/11/25 11:50:15 saving Message ID:  1674c3c1f389215b
2018/11/25 11:50:15 Indexed data id 1674c3c1f389215b to index mail, type document
```

### Deleting index
You can use `curl` to delete an email from the index using its id:
```bash
curl -XDELETE localhost:9200/mail/document/16752895eba45df6
```

Or you can delete the entire index:
```bash
curl -XDELETE localhost:9200/mail
```

# Unit tests
To run tests:

```bash
go test ./gmailservice
```