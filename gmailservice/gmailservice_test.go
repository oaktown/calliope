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
  "strconv"
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

  type results struct {
    failures  int
    successes int
  }

  type allResults struct {
    firstAdvance  results
    secondAdvance results
    finalAdvance  results
  }

  testWith429s := func(expected allResults, secsBeforeOk time.Duration) {
    var total429s, total200s int
    var messages, partialMessages []*store.Message
    startTimeStr := "2018-12-01 0:00:00 PST"
    maxWorkers := 3
    finalAdvance := RetryWaitInterval * time.Second * 10
    verify429sAnd200s := func(failures, successes int) {
      if total429s != failures {
        t.Errorf("Expected %v 429 responses. Instead, got %v", failures, total429s)
      }
      if total200s != successes {
        t.Errorf("Expected %v successful downloads. Instead, got %v", successes, total200s)
      }
    }

    failsFor := func(start time.Time, secs time.Duration) func(*Downloader, string) (*gmail.Message, error) {
      ok := start.Add(secs * time.Second)
      return func(d *Downloader, id string) (*gmail.Message, error) {
        now := d.clock.Now()
        if !now.Before(ok) {
          total200s++
          return fakeGmailMessage(id, "Successfully downloaded"), nil
        }
        total429s++
        return nil, err429
      }
    }

    // Test Downloader with time and network calls stubbed out
    layout := "2006-01-02 15:04:05 MST"
    start, _ := time.Parse(layout, startTimeStr)
    fakeClock := clockwork.NewFakeClockAt(start)
    downloader := New(nil, Options{}, maxWorkers)
    downloader.clock = fakeClock
    downloader.doGet = failsFor(start, secsBeforeOk)
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
      DownloadFullMessages(downloader)
      wg.Done()
    }()
    wg.Add(1)
    go func() {
      for m := range downloader.MessageChan {
        //fmt.Println("TTTT Subject:", m.Subject)
        if m.Subject == "Successfully downloaded" {
          messages = append(messages, m)
        } else {
          partialMessages = append(partialMessages, m)
        }
      }
      wg.Done()
    }()

    go func() {
      for i := 1; i <= maxWorkers; i++ {
        downloader.SearchChan <- fakeGmailMessage(strconv.Itoa(i), "")
      }
      close(downloader.SearchChan)
    }()
    // Wait until all three goroutines have called Sleep the first time
    fakeClock.BlockUntil(maxWorkers)
    verify429sAnd200s(expected.firstAdvance.failures, expected.firstAdvance.successes)

    // Advance fakeClock to first retry (still before the time we start responding w/ 200s)
    fakeClock.Advance(RetryWaitInterval * time.Second)
    // Wait until all three goroutines have called Sleep the second time
    fakeClock.BlockUntil(maxWorkers)
    verify429sAnd200s(expected.secondAdvance.failures, expected.secondAdvance.successes)

    // Advance fakeClock into the time we respond with 200s, and grab all messages before final verification
    fakeClock.Advance(finalAdvance)
    wg.Wait()
    verify429sAnd200s(expected.finalAdvance.failures, expected.finalAdvance.successes)
    if len(messages) != expected.finalAdvance.successes {
      t.Errorf("Expected %v successful downloads. Instead, got %v", expected.finalAdvance.successes, len(messages))
    }
  }

  // Kind of unnecessary since we only have one test, but leaving it this way to make it easy to
  // add additional test scenarios.
  tests := []struct {
    name   string
    verify func()
  }{
    {
      name: "Test incremental back off when API usage exceeded",
      verify: func() {
        testWith429s(
          allResults{
            firstAdvance: results{
              failures:  3,
              successes: 0,
            },
            secondAdvance: results{
              failures:  6,
              successes: 0,
            },
            finalAdvance: results{
              failures:  6,
              successes: 3,
            },
          }, 59,
        )
      },
    },
    {
      name: "Test unavailable API",
      verify: func() {
        testWith429s(
          allResults{
            firstAdvance: results{
              failures:  3,
              successes: 0,
            },
            secondAdvance: results{
              failures:  6,
              successes: 0,
            },
            finalAdvance: results{
              failures:  9,
              successes: 0,
            },
          }, 180,
        )
      },
    },
  }
  for _, tt := range tests {
    tt.verify()
  }
}
