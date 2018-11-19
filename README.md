# Calliope
Email query and visualization for observation and study of wild email monsters

![angry faced envelope with arms waving](images/email_monster.png)


## Setting up Go

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

# Setup a Google Cloud project
