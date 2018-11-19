package auth

import (
  "fmt"
  "log"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "os"
  "os/user"
  "path/filepath"

  "google.golang.org/api/gmail/v1"
  "golang.org/x/net/context"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/google"
)

func oauthConfig() (*oauth2.Config, error) {
  const Configfile = "client_secret.json"
  secret, err := ioutil.ReadFile(Configfile);
  if err != nil {
    log.Printf("could not read config file: %v\n", Configfile)
    return nil, err
  }

  config, err := google.ConfigFromJSON(secret, gmail.GmailReadonlyScope)
  if err != nil {
    return nil, err
  }

  return config, nil

}

// localTokenFilename generates credential file path
func localTokenFilename() (string, error) {
  const Directory = ".credentials"
  usr, err := user.Current()
  if err != nil {
    log.Printf("could't get os user, not able to cache token locally: %v\n", err)
    return "", err
  } else {
    tokenCacheDir := filepath.Join(usr.HomeDir, Directory)
    os.MkdirAll(tokenCacheDir, 0700)
    return filepath.Join(tokenCacheDir,"calliope.json"), nil
  }
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, t *oauth2.Token) {
  fmt.Printf("caching token locally: %s\n", file);

  err := ioutil.WriteFile(file, []byte(t.AccessToken), 0644)
  if err != nil {
    log.Printf("could not cache token locally (%v), you will need to log in again next time\n", err)
  }
}


// getTokenFromWeb uses a Google OAuth config to request an auth token.
// It returns the retrieved Token.
func tokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
  authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
  fmt.Println("")
  fmt.Println("====> Get ready to authenticate....")
  fmt.Println("")
  fmt.Printf("Open the link below in your browser. To give permission to view your email, click 'Allow'"+
    " then copy the code...\n\n%v\n", authURL)
  fmt.Println("")
  fmt.Print("Paste the code here:")

  var code string
  var tok *oauth2.Token
  var err error

  if _, err := fmt.Scan(&code); err != nil {
    log.Printf("Unable to read authorization code, err: %v", err)
    return nil, err
  }

  // Exchange converts an authorization code into a token.
  if tok, err = config.Exchange(oauth2.NoContext, code); err != nil {
    log.Printf("Unable to retrieve token from web code, err: %v", err)
    return nil, err
  }
  return tok, nil
}



// Client returns an OAuth http client
func Client(ctx context.Context) (*http.Client, error) {
  fmt.Printf("hello auth\n")

  var err error;
  var token oauth2.Token
  var t *oauth2.Token;

  c, err := oauthConfig()
  if err != nil {
    return nil, err
  }

  // ----------------------------
  // code below seems to work, but doesn't actually
  // Unable to retrieve messages:
  // oauth2: token expired and refresh token is not set
  // ----------------------------
  cacheFile, err := localTokenFilename()
  if false {
    if err == nil {
      fmt.Printf("checking local file: %s\n", cacheFile)
      if raw, err := ioutil.ReadFile(cacheFile); err == nil {
        fmt.Printf("got client from local file: %s\n", cacheFile)
        json.Unmarshal(raw, &token)
        t = &token;
      }
    }
  }

  if t == nil {
    fmt.Printf("need to get from Web\n")
    if t, err = tokenFromWeb(c); err != nil {
      return nil, err
    }
    saveToken(cacheFile, t)		// ignore error, ok not to have local file
  }

  client := c.Client(ctx, t);
  fmt.Printf("got client\n")

  return client, err;
}