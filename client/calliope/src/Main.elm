module Main exposing (ChartDay, Message, Model, Msg(..), RawSearchForm, RawSearchFormMsg(..), SearchForm, SearchFormMsg(..), SearchResults, init, main, searchResultsDecoder, subscriptions, timeSeries, update, updateRawSearchForm, updateSearchForm, view, viewRawSearchForm, viewSearchForm, viewSearchForms)

import BarGraph exposing (barGraph)
import Browser
import Browser.Events as BE
import Debug
import Element as E
    exposing
        ( Element
        , column
        , el
        , fill
        , height
        , padding
        , rgb255
        , row
        , width
        )
import Element.Background as Background
import Element.Border as Border
import Element.Events as Ev
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html, iframe)
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
    , showAdvancedSearch : Bool
    , windowWidth : Int
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


init : Int -> ( Model, Cmd Msg )
init flags =
    let
        defaultSearchForm =
            SearchForm "" "" "" "" "" "" False "Date" False 100

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
      , showAdvancedSearch = False
      , windowWidth = flags
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
    | CollapseAll
    | ToggleAdvancedSearch
    | Resize Int Int
    | GotSearch (Result Http.Error SearchResults)
    | Toggle String


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    let
        updateMessagesInSearchResults : (Message -> Message) -> SearchResults
        updateMessagesInSearchResults updateMessage =
            let
                messages =
                    List.map updateMessage model.searchResults.messages

                searchResults =
                    model.searchResults
            in
            { searchResults | messages = messages }
    in
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

        ToggleAdvancedSearch ->
            ( { model | showAdvancedSearch = not model.showAdvancedSearch }, Cmd.none )

        Toggle id ->
            let
                toggle message =
                    if message.id == id then
                        { message | expanded = not message.expanded }

                    else
                        message
            in
            ( { model | searchResults = updateMessagesInSearchResults toggle }, Cmd.none )

        CollapseAll ->
            let
                setCollapse message =
                    { message | expanded = False }
            in
            ( { model | searchResults = updateMessagesInSearchResults setCollapse }, Cmd.none )

        Resize x _ ->
            ( { model | windowWidth = x }, Cmd.none )

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
    BE.onResize (\x y -> Resize x y)



--    Sub.none
-- VIEW


view : Model -> Html Msg
view model =
    E.layout [ Font.size 12 ] <|
        column [ width fill ]
            [ viewTopbar
            , viewSearchForms model
            , viewSearchResults model.windowWidth model.searchStatus model.searchResults model.gmailUrl
            ]


appTitle : Element msg
appTitle =
    el
        [ Font.color (rgb255 255 255 255)
        , padding 20
        , Font.size 20
        ]
        (E.text "Calliope")


viewTopbar : Element msg
viewTopbar =
    row [ width fill, Background.color <| rgb255 92 99 118 ]
        [ appTitle ]


spacing =
    20


defaultButtonAttrs =
    [ Border.width 1
    , Border.color gray
    , Border.rounded 5
    , Font.size 12
    , E.padding 3
    ]


graphWidth =
    800



{- TODO: Currently set in barGraph not in this file. When we replace the
   placeholder barGraph with the new visualization we should make the dimensions
   settable.
-}


graphHeight =
    450


gutter =
    E.paddingXY 20 0


black =
    E.rgb255 0 0 0


white =
    E.rgb255 255 255 255


gray =
    E.rgb255 220 220 220


dimmedGray =
    E.rgba255 120 120 120 50


linkColor =
    E.rgb255 30 30 200


inputTextStyle =
    [ padding 6 ]


viewSearchForms : Model -> E.Element Msg
viewSearchForms model =
    let
        searchForm =
            if model.showAdvancedSearch then
                viewRawSearchForm model.rawSearchForm

            else
                viewSearchForm model.searchForm
    in
    E.column
        [ E.spacing spacing
        , padding 20
        ]
        [ E.el [] <|
            Input.text inputTextStyle
                { onChange = \str -> UpdateGmailUrl str
                , text = model.gmailUrl
                , placeholder = Nothing
                , label = Input.labelAbove [] (E.text "Gmail url")
                }
        , searchForm
        ]


onOffSwitch checked =
    let
        backgroundColor =
            if checked then
                black

            else
                white
    in
    E.el [ Border.solid, Border.color gray, Border.width 1, Font.color backgroundColor, E.width (E.px 10), Background.color backgroundColor ] (E.text " ")


viewSearchForm : SearchForm -> E.Element Msg
viewSearchForm model =
    let
        searchField : (String -> SearchFormMsg) -> String -> String -> E.Element Msg
        searchField msg val label =
            Input.text inputTextStyle
                { onChange = \str -> UpdateSearch (msg str)
                , text = val
                , placeholder = Nothing
                , label = Input.labelAbove [] (E.text label)
                }

        columnAttrs =
            [ E.spacing spacing, E.width (E.px 500), E.alignTop ]

        leftSide =
            E.column columnAttrs
                [ searchField Participants model.participants "Participants (applies to From, To, and CC)"
                , searchField StartDate model.startDate "Start date (\"YYYY-MM-DD\")"
                , searchField EndDate model.endDate "End date (\"YYYY-MM-DD\")"
                , searchField TimeZone model.timeZone "Time zone (e.g. -0800 for PST)"
                ]

        rightSide =
            E.column (gutter :: columnAttrs)
                [ searchField BodyOrSubject model.bodyOrSubject "Body or subject"
                , E.row [ E.width E.fill ]
                    [ Input.text (E.width (E.fillPortion 4) :: inputTextStyle)
                        { onChange = \str -> UpdateSearch (Label str)
                        , text = model.label
                        , placeholder = Nothing
                        , label = Input.labelAbove [] (E.text "Label")
                        }
                    , Input.checkbox [ E.width (E.fillPortion 1), gutter ]
                        { onChange = \b -> UpdateSearch StarredOnly
                        , icon = onOffSwitch
                        , checked = model.starredOnly
                        , label = Input.labelLeft [] (E.text "Starred only")
                        }
                    ]
                , E.row [ E.width E.fill ]
                    [ Input.text (E.width (E.fillPortion 4) :: inputTextStyle)
                        { onChange = \str -> UpdateSearch (SortField str)
                        , text = model.sortField
                        , placeholder = Nothing
                        , label = Input.labelAbove [] (E.text "Sort field")
                        }
                    , Input.checkbox [ E.width (E.fillPortion 1), gutter ]
                        { onChange = \b -> UpdateSearch Ascending
                        , icon = onOffSwitch
                        , checked = model.ascending
                        , label = Input.labelLeft [] (E.text "Ascending")
                        }
                    ]
                , searchField Size (String.fromInt model.size) "Size"
                ]

        formFields =
            E.row [] [ leftSide, rightSide ]

        buttons =
            E.row []
                [ Input.button defaultButtonAttrs
                    { onPress = Just DoSearch
                    , label = E.text "Query"
                    }
                , Input.button defaultButtonAttrs
                    { onPress = Just ToggleAdvancedSearch
                    , label = E.text "AdvancedSearch"
                    }
                ]
    in
    E.row []
        [ E.column [] [ formFields, buttons ]
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
        , E.row []
            [ Input.button defaultButtonAttrs
                { onPress = Just DoRawSearch
                , label = E.text "Raw query"
                }
            , Input.button defaultButtonAttrs
                { onPress = Just ToggleAdvancedSearch
                , label = E.text "Regular search"
                }
            ]
        ]


viewSearchResults : Int -> SearchStatus -> SearchResults -> String -> E.Element Msg
viewSearchResults windowWidth status searchResults inboxUrl =
    let
        expandAndCollapseButtons =
            E.row []
                [ Input.button defaultButtonAttrs
                    { onPress = Just CollapseAll
                    , label = E.text "Collapse all"
                    }
                ]

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
                                            , E.el [ Font.color dimmedGray ] <| E.text (" – " ++ message.snippet)
                                            ]
                                        ]

                                expanded =
                                    let
                                        iframe =
                                            Html.iframe [ Attributes.width windowWidth, Attributes.height graphHeight, Attributes.src ("/message/" ++ message.id) ] []
                                    in
                                    if message.expanded then
                                        E.column []
                                            [ E.link [ Font.color linkColor ]
                                                { url = threadUrl message.threadId
                                                , label = E.text "Open in Gmail"
                                                }
                                            , E.el [ E.width E.fill ] (E.html iframe)
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
                    [ E.el [ E.width (E.px graphWidth), E.height E.fill ] (E.html <| Html.div [ Attributes.id "graph" ] [ barGraph (timeSeries searchResults.chartData) ])
                    , expandAndCollapseButtons
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
