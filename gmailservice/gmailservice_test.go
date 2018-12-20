package gmailservice

import (
  "encoding/json"
  "github.com/jonboulle/clockwork"
  "github.com/oaktown/calliope/store"
  "google.golang.org/api/googleapi"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
  "sync"
  "testing"
  "time"

  "google.golang.org/api/gmail/v1"
)

var expectedBody = "How are you doing? Some special characters: <>?$/:-)"

func getEmailJson() []byte {
  jsonFile, err := os.Open("gmail_message.json") // Slack notification

  if err != nil {
    log.Fatalf("Couldn't open gmail_message.json. Error: %v", err)
  }

  emailJson, _ := ioutil.ReadAll(jsonFile) // theoretically the same thing that the gmail API returns
  jsonFile.Close()
  return emailJson
}

func JsonToGmail(jsonByteArray []byte) (gmail.Message, error) {
  var data gmail.Message
  if err := json.Unmarshal(jsonByteArray, &data); err != nil {
    log.Printf("json.Unmarshal failed, skipping message, err: %v", err)
    return data, err
  }
  return data, nil
}

func TestBodyText(t *testing.T) {
  emailJson := getEmailJson()
  gmailMsg, _ := JsonToGmail(emailJson)

  if body := BodyText(gmailMsg); !strings.Contains(body, expectedBody) {
    t.Errorf("Body is incorrect. Should have been:\n\n%v\n\nInstead, got:\n\n%v\n\n", expectedBody, body)
  }
}

func TestGmailToMessage(t *testing.T) {
  emailJson := getEmailJson()
  rawGmail, _ := JsonToGmail(emailJson)
  msg, err := GmailToMessage(rawGmail, "https://mail.google.com/mail/#inbox", time.Now())
  if err != nil {
    t.Errorf("Unexpected error calling JsonForElasticsearch: %v", err)
  }
  if body := msg.Body; !strings.Contains(body, expectedBody) {
    t.Errorf("Body is incorrect. Should have been:\n\n%v\n\nInstead, got:\n\n%v\n\n", expectedBody, body)
  }
}

func TestDownloadFullMessages(t *testing.T) {

  fakeGmailMessage := func(id, subject string) *gmail.Message {
    return &gmail.Message{
      Id: id,
      Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
          {
            Name:  "Subject",
            Value: subject,
          },
        },
        Body: &gmail.MessagePartBody{
          Data: "",
        },
      },
    }
  }

  err429 := &googleapi.Error{
    Code:   429,
    Header: make(http.Header),
  }

  tooManyRequests429 := func() {
    var total429s int

    failsFor := func(start time.Time, secs time.Duration) func(*Downloader, string) (*gmail.Message, error) {
      ok := start.Add(time.Duration(secs) * time.Second)
      return func(d *Downloader, id string) (*gmail.Message, error) {
        if !d.clock.Now().Before(ok) {
          return fakeGmailMessage(id, "Successfully downloaded"), nil
        }
        total429s++
        return nil, err429
      }
    }

    // Test Downloader with time and network calls stubbed out
    layout := "2006-01-02 15:04:05 MST"
    start, _ := time.Parse(layout, "2018-12-01 0:00:00 PST")
    fakeClock := clockwork.NewFakeClockAt(start)
    downloader := New(nil, Options{}, 3)
    downloader.clock = fakeClock
    downloader.doGet = failsFor(start, 59)
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
      DownloadFullMessages(downloader)
      wg.Done()
    }()
    downloader.SearchChan <- fakeGmailMessage("1", "")
    downloader.SearchChan <- fakeGmailMessage("2", "")
    downloader.SearchChan <- fakeGmailMessage("3", "")
    close(downloader.SearchChan)

    // Wait until all three goroutines have called Sleep the first time
    fakeClock.BlockUntil(3)

    if total429s != 3 {
      t.Errorf("Expected 3 429 responses. Instead, got %v", total429s)
    }

    if len(downloader.MessageChan) > 0 {
      t.Errorf("Expected 0 downloads (because of 429s). Instead, got %v", len(downloader.MessageChan))
    }

    // Advance fakeClock to first retry (still before the time we start responding w/ 200s)
    fakeClock.Advance(30 * time.Second)

    // Wait until all three goroutines have called Sleep the second time
    fakeClock.BlockUntil(3)

    if total429s != 6 {
      t.Errorf("Expected 6 429 responses. Instead, got %v", total429s)
    }

    // Advance fakeClock into the time we respond with 200
    fakeClock.Advance(120 * time.Second)
    var messages []*store.Message
    for m := range downloader.MessageChan {
      messages = append(messages, m)
    }

    wg.Wait()
    if len(messages) != 3 {
      t.Errorf("Expected 3 successful downloads. Instead, got %v", len(messages))
    }
  }

  // Kind of unnecessary since we only have one test, but leaving it this way to make it easy to
  // add additional test scenarios.
  tests := []struct {
    name   string
    verify func()
  }{
    {
      name:   "Test incremental back off when API usage exceeded",
      verify: tooManyRequests429,
    },
  }
  for _, tt := range tests {
    tt.verify()
  }
}
