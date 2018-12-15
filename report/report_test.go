package report

import (
  "github.com/oaktown/calliope/store"
  "reflect"
  "testing"
  "time"
)

func m(s string) *store.Message {
  date, _ := time.Parse("2006/01/02", s)
  return &store.Message{
    Date: date,
  }
}

func Test_getChartData(t *testing.T) {
  type args struct {
    messages []*store.Message
  }

  tests := []struct {
    name string
    args args
    want []BarData
  }{
    {
      name: "Contiguous ordered",
      args: args{
        messages: []*store.Message{
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
    {
      name: "Contiguous out of order",
      args: args{
        messages: []*store.Message{
          m("2018/01/03"),
          m("2018/01/01"),
          m("2018/01/02"),
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
    {
      name: "Non-contiguous",
      args: args{
        messages: []*store.Message{
          m("2018/01/01"),
          m("2018/01/02"),
          m("2018/01/05"),
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
          Messages: 0,
        },
        {
          Date:     "2018-01-04",
          Messages: 0,
        },
        {
          Date:     "2018-01-05",
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
