# Calliope
Email query and visualization for observation and study of wild email monsters

![angry faced envelope with arms waving](images/email_monster.png)

## Setting up Development Environment

Calliope's backend is written in Go. This includes the CLI and an API server. The front-end is written in Elm. The following instructions are for developing with Macs.

Install go using homebrew:

```
brew install go
brew install dep # Dependency management
```

Go requires your project directory to be a subdirectory of `$GOPATH/src`. Go expects all
of your go projects to be under this path. To get the code and put it in the right path, use the `go get` command (instead of `git clone`)

```bash
go get github.com/oaktown/calliope.git
cd $GOPATH
cd src/github.com/oaktown/calliope # Project directory
dep ensure # Update dependencies
```
The last line shouldn't cause any changes since we're checking all dependencies into `vendor` (because, unlike npm or ruby gems, there is no registry – if owners rename repos, it would cause a problem for anyone using their packages.)

### Install Elasticsearch using Docker

This is adapted from olivere's [elastic-with-docker repo](https://github.com/olivere/elastic-with-docker).

_NOTE: The following should be done from the project directory._

Copy the example Docker Compose file:

```bash
cp docker-compose.example.yml docker-compose.yml
```

Edit it to your needs. Then start up Elasticsearch:


```bash
docker-compose up
```

This will download the docker images (if you don't have them already) and run Elasticsearch on port 9200. Note: The command above will run in the foreground where you will see logging output. You can stop the container with `ctrl-c`. If you want to run in the background instead, attach the `-d` flag.


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

If you want data to persist (from [elastic-with-docker repo](https://github.com/olivere/elastic-with-docker)):

> Make sure to create a ./data directory locally and uncomment the volumes section in Docker Compose file(s) if you want your data to be persistent.

Note: `data` has been added to .gitignore

To delete your containers (e.g. if you don't want to save what's in Elasticsearch and haven't put stuff in the data directory), type:

```bash
docker-compose -f docker-compose.local.yml down
```

There's a Chrome extension called [ElasticSearch Head](https://chrome.google.com/webstore/detail/elasticsearch-head/ffmkiejjmecolpfloofpjologoblkegm) that you might find useful.

### Setup a Google Cloud project

We were inspired by https://github.com/vcollak/GmailContacts, which has nice
setup instructions.

### Setup Elm

We are using Elm 0.19. Install Elm using homebrew and [create-elm-app](https://github.com/halfzebra/create-elm-app) and [elm-format](https://github.com/avh4/elm-format) using npm (or alternative, like yarn):
```bash
brew install elm
elm --version # Should return 0.19 or higher
npm install -g create-elm-app # Make sure you are running Node version >=8
npm install -g elm-format
```

Elm code is in `client/calliope`.

We use `create-elm-app` to serve the app in developer mode (including proxying to the API server). It also can create an optimized build for production.

We use `elm-format` to format code. It's recommended you use an editor/IDE that can call `elm-format` for you, such as:
- [IntelliJ](https://plugins.jetbrains.com/plugin/10268-elm) (including GoLand). You'll want to follow the instructions on [elm-format](https://github.com/klazuka/intellij-elm/blob/master/docs/elm-format/setup.md) (this will be automated in the future)
- [Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=sbrink.elm)

Additional configuration may be necessary to use `elm-format` with the above.

## Build and run the app

When building the app `-i` will build all the dependencies and `-v` prints
out what its building.  The following commands will generate an executable
called `calliope` in the same directory and then run it:

```bash
go build -i -v
./calliope
```
or you can use `go run`:

```bash
go run main.go
```

This command will display help.

Note: One of the options for several commands is an url to open the thread in Gmail (not the specific email, just the thread that it's in).
By default, the url is `https://mail.google.com/mail/#inbox/<thread id>`, but if you are logged into more than one
account, can pass in a modified url. For example:

```bash
go build
./calliope download -l 1000 -d "2018/01/01" -u "https://mail.google.com/mail/u/2"
```

or

```bash
go run main.go download -l 1000 -d "2018/01/01" -u "https://mail.google.com/mail/u/2/"
```

will link to the 3rd logged in account. See also [Debugging](#debugging), below.

After you've downloaded messages, you can run the server:

```bash
go run main.go web
```

This should start listening on port 8080.

Now start the front-end:

```bash
cd client/calliope
elm-app start
```

You should then be able to visit the app on port 3000.

Test to make sure that the development web app is proxying correctly to the backend:

```bash
 curl -X GET -H "Content-type: application/json" -H "Accept: application/json"  http://localhost:3000/api/search

``` 

**Note:** `create-elm-app` starts a web server that auto-compiles your Elm code. It supports hot-module reloading, so as you make changes to the code, it should be reflected in the app. It connects the Elm app to the API server by proxying. See [create-elm-app docs](https://github.com/halfzebra/create-elm-app/blob/master/template/README.md#setting-up-api-proxy) for more information.

### Config

In addition to command line options, you can provide a configuration file (currently named calliope.yml, and currently stored in the working directory, although this will change eventually). Currently, it only has one option: `exclude_headers_with_values` which can be used to exclude messages from being saved into Elasticsearch (useful if you want to filter out automated notifications, email lists, etc.). There is a sample file `calliope-example.yml` that shows a configuration to exclude common mailing lists.

### Oauth

The first time you run the application, you will be prompted to give permission (via Oauth) like so:

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

### Deleting index
You can use `curl` to delete an email from the index using its id:
```bash
curl -XDELETE localhost:9200/mail/document/<id>
```

Or you can delete the entire index:
```bash
curl -XDELETE localhost:9200/mail
```

## Unit tests
We have unit tests for both Go and Elm. 

### Go unit tests
To run Go unit tests:

```bash
go test -v ./...
```

The `-v` option is optional; it provides more info and also ignores test cache. The `./..`
means to run all tests. Alternatively, you can provide a path to the package (e.g.
`./gmailservice`).

Note: If you find a problematic email and want to download it locally e.g. to add as a test
fixture, you can get it using curl:

```bash
curl localhost:9200/mail/document/<id> > fixture.json
```

Removing stuff you don't need for the test would be nice, too, as it would make it easier to
find relevant data in the fixture.

### Elm unit tests
To run Elm unit tests:
```bash
cd client/calliope
elm-app test
```


## Debugging

### Debugging Go   
To debug, install [Delve](https://github.com/derekparker/delve). Follow the installation
instructions for your OS, then you can run it like:

```bash
dlv debug main.go -- download -q "is:starred label:devchix" -l 10 -R
```

Note: The `--` separates `dlv` commandline args from the commandline args of the program being debugged.

To debug tests that have fixtures, you will need to change the working directory:

```bash
cd gmailservice
dlv test . -- -test.v
```

If you don't change working directories, it'll have trouble finding fixture. Also, `./..`
from the project root doesn't work.

### Debugging Elm

When you run the app using `elm-app start`, you can access the debugger through the `Explore History` widget at the bottom. More info on the debugger in this [blog post from Elm creator Evan Czaplicki](https://elm-lang.org/blog/the-perfect-bug-report).

## Contributing

All changes via Pull Request, someone who is not the author should provide
feedback and approve changes.

If the author is a committer, then leave the PR merge to them.  Committers
should rebase and squash before merging to master.




