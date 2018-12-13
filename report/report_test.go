package report

import (
  "reflect"
  "testing"
  "time"

  "github.com/oaktown/calliope/gmailservice"
)

func m(s string) *gmailservice.Message {
  date, _ := time.Parse("2006/01/02", s)
  return &gmailservice.Message{
    Date: date,
  }
}

func Test_getChartData(t *testing.T) {
  type args struct {
    messages []*gmailservice.Message
  }

  tests := []struct {
    name string
    args args
    want []BarData
  }{
    {
      name: "Basic",
      args: args{
        messages: []*gmailservice.Message{
          m("2018/01/01"),
          m("2018/01/02"),
          m("2018/01/03"),
          m("2018/01/01"),
          m("2018/01/01"),
          m("2018/01/02"),
        },
      },
      want: []BarData{
        {
          Date:     "2018-01-01",
          Messages: 3,
        },
        {
          Date:     "2018-01-02",
          Messages: 2,
        },
        {
          Date:     "2018-01-03",
          Messages: 1,
        },
      },
    },
  }
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      if got := getChartData(tt.args.messages); !reflect.DeepEqual(got, tt.want) {
        t.Errorf("getChartData() = %v, want %v", got, tt.want)
      }
    })
  }
}
