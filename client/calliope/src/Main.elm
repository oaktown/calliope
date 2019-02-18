module Main exposing (ChartDay, Message, Model, Msg(..), RawSearchForm, RawSearchFormMsg(..), SearchForm, SearchFormMsg(..), SearchResults, init, main, searchResultsDecoder, subscriptions, timeSeries, update, updateRawSearchForm, updateSearchForm, view, viewRawSearchForm, viewSearchForm, viewSearchForms)

import BarGraph exposing (barGraph)
import Browser
import Debug
import Element as E
import Element.Background as Background
import Element.Border as Border
import Element.Events as Ev
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html)
import Html.Attributes as Attributes
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder, field, int, list, string)
import Json.Decode.Pipeline exposing (hardcoded, required)
import Time
import Url.Builder



-- MAIN


main =
    Browser.element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }



-- MODEL


type alias SearchForm =
    { participants : String
    , bodyOrSubject : String
    , startDate : String
    , endDate : String
    , timeZone : String
    , label : String
    , starredOnly : Bool
    , sortField : String
    , ascending : Bool
    , size : Int
    }


type alias RawSearchForm =
    { query : String
    }


type alias Model =
    { gmailUrl : String
    , searchForm : SearchForm
    , rawSearchForm : RawSearchForm
    , searchResults : SearchResults
    , searchStatus : SearchStatus
    }


type alias Message =
    { id : String
    , threadId : String
    , labelIds : List String
    , date : String
    , downloadedStartedAt : String
    , to : String
    , cc : String
    , from : String
    , subject : String
    , snippet : String
    , body : String
    , expanded : Bool
    }


type alias ChartDay =
    { date : String
    , messages : Int
    }


type SearchStatus
    = Empty
    | Loading
    | Success
    | Failure String


type alias SearchResults =
    { query : String
    , chartData : List ChartDay
    , messages : List Message
    }


init : () -> ( Model, Cmd Msg )
init _ =
    let
        defaultSearchForm =
            SearchForm "" "" "" "" "" "" False "" False 100

        defaultRawSearchForm =
            RawSearchForm """{
  "query": {
    "match_all": {}
  },
  "size": 100,
  "sort": [
    {
      "Date": {
        "order": "desc"
      }
    }
  ]
}"""

        emptySearchResults =
            SearchResults "" [] []
    in
    ( { gmailUrl = "https://mail.google.com/mail/"
      , searchForm = defaultSearchForm
      , rawSearchForm = defaultRawSearchForm
      , searchResults = emptySearchResults
      , searchStatus = Empty
      }
    , Cmd.none
    )



-- UPDATE


type RawSearchFormMsg
    = Query String


type SearchFormMsg
    = Participants String
    | BodyOrSubject String
    | StartDate String
    | EndDate String
    | TimeZone String
    | Label String
    | StarredOnly
    | SortField String
    | Ascending
    | Size String


type Msg
    = UpdateGmailUrl String
    | UpdateRawSearch RawSearchFormMsg
    | UpdateSearch SearchFormMsg
    | DoSearch
    | DoRawSearch
    | GotSearch (Result Http.Error SearchResults)
    | Toggle String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UpdateGmailUrl gmailUrl ->
            ( { model | gmailUrl = gmailUrl }, Cmd.none )

        UpdateRawSearch field ->
            ( { model
                | rawSearchForm = updateRawSearchForm field model.rawSearchForm
              }
            , Cmd.none
            )

        UpdateSearch field ->
            ( { model
                | searchForm = updateSearchForm field model.searchForm
              }
            , Cmd.none
            )

        Toggle id ->
            let
                toggle message =
                    if message.id == id then
                        { message | expanded = not message.expanded }

                    else
                        message

                messages =
                    List.map toggle model.searchResults.messages

                searchResults =
                    let
                        oldSearchResults =
                            model.searchResults
                    in
                    { oldSearchResults | messages = messages }
            in
            ( { model | searchResults = searchResults }, Cmd.none )

        DoSearch ->
            ( { model | searchStatus = Loading }, doSearch model.searchForm )

        DoRawSearch ->
            ( { model | searchStatus = Loading }, doRawSearch model.rawSearchForm )

        GotSearch results ->
            case results of
                Ok searchResults ->
                    let
                        rawSearchForm =
                            model.rawSearchForm

                        updatedRawSearchForm =
                            { rawSearchForm | query = searchResults.query }
                    in
                    ( { model
                        | searchResults = searchResults
                        , rawSearchForm = updatedRawSearchForm
                        , searchStatus = Success
                      }
                    , Cmd.none
                    )

                Err e ->
                    ( { model | searchStatus = Failure (Debug.toString e) }, Cmd.none )


updateRawSearchForm : RawSearchFormMsg -> RawSearchForm -> RawSearchForm
updateRawSearchForm msg model =
    case msg of
        Query query ->
            { model | query = query }


updateSearchForm : SearchFormMsg -> SearchForm -> SearchForm
updateSearchForm msg model =
    case msg of
        Participants participants ->
            { model | participants = participants }

        BodyOrSubject bodyOrSubject ->
            { model | bodyOrSubject = bodyOrSubject }

        StartDate startDate ->
            { model | startDate = startDate }

        EndDate endDate ->
            { model | endDate = endDate }

        TimeZone timeZone ->
            { model | timeZone = timeZone }

        Label label ->
            { model | label = label }

        StarredOnly ->
            { model | starredOnly = not model.starredOnly }

        SortField sortField ->
            { model | sortField = sortField }

        Ascending ->
            { model | ascending = not model.ascending }

        Size size ->
            case String.toInt size of
                Just int ->
                    { model | size = int }

                Nothing ->
                    { model | size = 0 }



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    E.layout [ Font.size 12 ] <|
        E.column [] [ viewSearchForms model, viewSearchResults model.searchStatus model.searchResults model.gmailUrl ]


spacing =
    20


viewSearchForms : Model -> E.Element Msg
viewSearchForms model =
    E.column [ E.spacing spacing ]
        [ E.el [] <|
            Input.text []
                { onChange = \str -> UpdateGmailUrl str
                , text = model.gmailUrl
                , placeholder = Nothing
                , label = Input.labelAbove [] (E.text "Gmail url:")
                }
        , viewSearchForm model.searchForm
        , E.el [] <| E.paragraph [] [ E.text "OR" ]
        , viewRawSearchForm model.rawSearchForm
        ]


onOffSwitch checked =
    let
        backgroundColor =
            if checked then
                E.rgb255 0 0 0

            else
                E.rgb255 255 255 255
    in
    E.el [ Border.solid, Border.color (E.rgb255 200 200 200), Border.width 1, Font.color backgroundColor, E.width (E.px 10), Background.color backgroundColor ] (E.text " ")


viewSearchForm : SearchForm -> E.Element Msg
viewSearchForm model =
    E.column [ E.spacing spacing ]
        [ Input.text []
            { onChange = \str -> UpdateSearch (Participants str)
            , text = model.participants
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Participants (applies to From:, To:, and CC:):")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (BodyOrSubject str)
            , text = model.bodyOrSubject
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Body or subject")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (StartDate str)
            , text = model.startDate
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Start date (\"YYYY-MM-DD\"):")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (EndDate str)
            , text = model.endDate
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "End date (\"YYYY-MM-DD\"):")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (TimeZone str)
            , text = model.timeZone
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Time zone (e.g. -0800 for PST):")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (Label str)
            , text = model.label
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Label:")
            }
        , Input.checkbox []
            { onChange = \b -> UpdateSearch StarredOnly
            , icon = onOffSwitch
            , checked = model.starredOnly
            , label = Input.labelLeft [] (E.text "Starred only:")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (SortField str)
            , text = model.sortField
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Sort field:")
            }
        , Input.checkbox []
            { onChange = \b -> UpdateSearch Ascending
            , icon = onOffSwitch
            , checked = model.ascending
            , label = Input.labelLeft [] (E.text "Ascending:")
            }
        , Input.text []
            { onChange = \str -> UpdateSearch (Size str)
            , text = String.fromInt model.size
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Size:")
            }
        , Input.button
            [ Border.width 1
            , Border.color <| E.rgb255 220 220 220
            , Border.rounded 5
            , Font.size 12
            , E.padding 3
            ]
            { onPress = Just DoSearch
            , label = E.text "Query"
            }
        ]


viewRawSearchForm : RawSearchForm -> E.Element Msg
viewRawSearchForm model =
    E.column
        []
        [ Input.multiline
            [ E.height E.shrink ]
            { onChange = \str -> UpdateRawSearch (Query str)
            , text = model.query
            , placeholder = Nothing
            , label = Input.labelAbove [] (E.text "Query")
            , spellcheck = False
            }
        , Input.button
            [ Border.width 1
            , Border.color <| E.rgb255 220 220 220
            , Border.rounded 5
            , Font.size 12
            , E.padding 3
            ]
            { onPress = Just DoRawSearch
            , label = E.text "Raw query"
            }
        ]


viewSearchResults : SearchStatus -> SearchResults -> String -> E.Element Msg
viewSearchResults status searchResults inboxUrl =
    let
        threadUrl =
            \id ->
                inboxUrl ++ "#inbox/" ++ id

        messageSummaries : List Message -> E.Element Msg
        messageSummaries messages =
            let
                messageRows : List (E.Element Msg)
                messageRows =
                    let
                        row : Message -> E.Element Msg
                        row message =
                            let
                                summary =
                                    E.row [ E.spacingXY 20 5, Ev.onClick (Toggle message.id) ]
                                        [ E.el [ E.width (E.px 280) ] (E.text message.date)
                                        , E.el [ E.width (E.px 300), E.clip ] (E.text message.from)
                                        , E.row []
                                            [ E.el [] <| E.text message.subject
                                            , E.el [ Font.color <| E.rgba255 120 120 120 50 ] <| E.text <| " – " ++ message.snippet
                                            ]
                                        ]

                                expanded =
                                    if message.expanded then
                                        E.column []
                                            [ E.link [ Font.color (E.rgb255 30 30 200) ]
                                                { url = threadUrl message.threadId
                                                , label = E.text "Open in Gmail"
                                                }
                                            , E.el [] (E.text message.body)
                                            ]

                                    else
                                        E.none
                            in
                            E.column [] [ summary, expanded ]
                    in
                    List.map row messages
            in
            E.column [] messageRows
    in
    case status of
        Loading ->
            E.el [] (E.text "Loading …")

        Success ->
            if List.length searchResults.messages > 0 then
                E.column []
                    [ E.el [ E.width (E.px 800), E.height E.fill ] (E.html <| Html.div [ Attributes.id "graph" ] [ barGraph (timeSeries searchResults.chartData) ])
                    , messageSummaries searchResults.messages
                    ]

            else
                E.el [] (E.text "No messages found")

        Failure e ->
            E.el [] <| E.text ("Search error:" ++ e)

        Empty ->
            E.none


timeSeries : List ChartDay -> List ( Time.Posix, Float )
timeSeries data =
    let
        convert : ChartDay -> ( Time.Posix, Float )
        convert t =
            let
                date =
                    Iso8601.toTime t.date
            in
            case date of
                Ok d ->
                    ( d, toFloat t.messages )

                -- need a better way of dealing with this
                Err _ ->
                    ( Time.millisToPosix 0, toFloat t.messages )
    in
    List.map convert data



-- HTTP


doSearch : SearchForm -> Cmd Msg
doSearch searchForm =
    let
        string =
            Url.Builder.string

        int =
            Url.Builder.int

        bool =
            \name b ->
                if b then
                    string name "true"

                else
                    string name ""

        params =
            [ string "participants" searchForm.participants
            , string "bodyOrSubject" searchForm.bodyOrSubject
            , string "startDate" searchForm.startDate
            , string "endDate" searchForm.endDate
            , string "timeZone" searchForm.timeZone
            , string "label" searchForm.label
            , bool "starredOnly" searchForm.starredOnly
            , string "sortField" searchForm.sortField
            , bool "ascending" searchForm.ascending
            , int "size" searchForm.size
            ]

        url =
            Url.Builder.absolute [ "api", "search" ] params
    in
    Http.get
        { url = url
        , expect = Http.expectJson GotSearch searchResultsDecoder
        }


doRawSearch : RawSearchForm -> Cmd Msg
doRawSearch rawSearchForm =
    let
        string =
            Url.Builder.string

        url =
            Url.Builder.absolute [ "api", "search" ] [ string "query" rawSearchForm.query ]
    in
    Http.get
        { url = url
        , expect = Http.expectJson GotSearch searchResultsDecoder
        }



-- Decoders


messageDecoder : Decoder Message
messageDecoder =
    Decode.succeed Message
        |> required "Id" string
        |> required "ThreadId" string
        |> required "LabelIds" (list string)
        |> required "Date" string
        |> required "DownloadedStartedAt" string
        |> required "To" string
        |> required "Cc" string
        |> required "From" string
        |> required "Subject" string
        |> required "Snippet" string
        |> required "Body" string
        |> hardcoded False


chartDayDecoder : Decoder ChartDay
chartDayDecoder =
    Decode.succeed ChartDay
        |> required "Date" string
        |> required "Messages" int


searchResultsDecoder : Decoder SearchResults
searchResultsDecoder =
    Decode.succeed SearchResults
        |> required "Query" string
        |> required "ChartData" (list chartDayDecoder)
        |> required "Messages" (list messageDecoder)
