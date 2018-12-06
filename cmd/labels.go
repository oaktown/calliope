package cmd

import (
  "fmt"
  "github.com/oaktown/calliope/misc"
  "github.com/oaktown/calliope/store"
  "github.com/spf13/cobra"
)

var labelName string

func init() {
  rootCmd.AddCommand(labelsCmd)
  labelsCmd.Flags().StringVarP(&labelName, "label", "l", "", "Look up the label id of a particular label.")

}

var labelsCmd = &cobra.Command{
  Use:   "labels",
  Short: "Get labels and ids from Elasticsearch",
  Long:  `Get labels and ids from Elasticsearch. Must have downloaded earlier.`,
  Run: func(cmd *cobra.Command, args []string) {
    store := misc.GetStoreClient()
    switch labelName == ""{
    case false:
      lookupLabel(store)
    default:
      showLabels(store)
    }
  },
}

func lookupLabel(store *store.Service) {
  labelId, err := store.FindLabelId(labelName)
  if err != nil {
    fmt.Printf("Error looking up label %v: %v\n", labelName, err)
  }
  fmt.Println(labelId)
}

func showLabels(store *store.Service) {
  labels, err := store.GetLabels()
  if err != nil {
    fmt.Println("Could not get labels from Elasticsearch. Error: ", err)
  }
  for _, label := range labels {

    fmt.Printf("|%30s|%-30s|\n", label.Name, label.Id)
  }
}
