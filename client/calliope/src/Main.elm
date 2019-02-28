module Main exposing (main)

import BarGraph exposing (barGraph)
import Browser
import Browser.Events
import Debug
import Element exposing (Element, alignTop, clip, column, el, fill, fillPortion, height, html, layout, link, none, padding, paddingXY, px, rgb255, rgba255, row, shrink, spacing, spacingXY, text, width)
import Element.Background as Background
import Element.Border as Border
import Element.Events as Events
import Element.Font as Font
import Element.Input as Input
import Html exposing (Html, iframe)
import Html.Attributes as Attributes
import Html.Parser
import Html.Parser.Util
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
    , bodyHtml : String
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
                        { message | expanded = False }
            in
            ( { model | searchResults = updateMessagesInSearchResults toggle }, Cmd.none )

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
    Browser.Events.onResize (\x y -> Resize x y)



--    Sub.none
-- VIEW


view : Model -> Html Msg
view model =
    layout [ Font.size 12 ] <|
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
        (text "Calliope")


viewTopbar : Element msg
viewTopbar =
    row [ width fill, Background.color <| rgb255 92 99 118 ]
        [ appTitle ]


space =
    20


defaultButtonAttrs =
    [ Border.width 1
    , Border.color gray
    , Border.rounded 5
    , Font.size 12
    , padding 3
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
    paddingXY 20 0


black =
    rgb255 0 0 0


white =
    rgb255 255 255 255


gray =
    rgb255 220 220 220


dimmedGray =
    rgba255 120 120 120 50


linkColor =
    rgb255 30 30 200


inputTextStyle =
    [ padding 6 ]


viewSearchForms : Model -> Element Msg
viewSearchForms model =
    let
        searchForm =
            if model.showAdvancedSearch then
                viewRawSearchForm model.rawSearchForm

            else
                viewSearchForm model.searchForm
    in
    column
        [ spacing space
        , padding 20
        ]
        [ el [ width fill ] <|
            Input.text inputTextStyle
                { onChange = \str -> UpdateGmailUrl str
                , text = model.gmailUrl
                , placeholder = Nothing
                , label = Input.labelAbove [] (text "Gmail url (useful if you are signed in to multiple Gmail accounts simultaneously)")
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
    el [ Border.solid, Border.color gray, Border.width 1, Font.color backgroundColor, width (px 10), Background.color backgroundColor ] (text " ")


viewSearchForm : SearchForm -> Element Msg
viewSearchForm model =
    let
        searchField : (String -> SearchFormMsg) -> String -> String -> Element Msg
        searchField msg val label =
            Input.text inputTextStyle
                { onChange = \str -> UpdateSearch (msg str)
                , text = val
                , placeholder = Nothing
                , label = Input.labelAbove [] (text label)
                }

        columnAttrs =
            [ spacing space, width (px 500), alignTop ]

        leftSide =
            column columnAttrs
                [ searchField Participants model.participants "Participants (applies to From, To, and CC)"
                , searchField StartDate model.startDate "Start date (\"YYYY-MM-DD\")"
                , searchField EndDate model.endDate "End date (\"YYYY-MM-DD\")"
                , searchField TimeZone model.timeZone "Time zone (e.g. -0800 for PST)"
                ]

        rightSide =
            column (gutter :: columnAttrs)
                [ searchField BodyOrSubject model.bodyOrSubject "Body or subject"
                , row [ width fill ]
                    [ Input.text (width (fillPortion 4) :: inputTextStyle)
                        { onChange = \str -> UpdateSearch (Label str)
                        , text = model.label
                        , placeholder = Nothing
                        , label = Input.labelAbove [] (text "Label")
                        }
                    , Input.checkbox [ width (fillPortion 1), gutter ]
                        { onChange = \b -> UpdateSearch StarredOnly
                        , icon = onOffSwitch
                        , checked = model.starredOnly
                        , label = Input.labelLeft [] (text "Starred only")
                        }
                    ]
                , row [ width fill ]
                    [ Input.text (width (fillPortion 4) :: inputTextStyle)
                        { onChange = \str -> UpdateSearch (SortField str)
                        , text = model.sortField
                        , placeholder = Nothing
                        , label = Input.labelAbove [] (text "Sort field")
                        }
                    , Input.checkbox [ width (fillPortion 1), gutter ]
                        { onChange = \b -> UpdateSearch Ascending
                        , icon = onOffSwitch
                        , checked = model.ascending
                        , label = Input.labelLeft [] (text "Ascending")
                        }
                    ]
                , searchField Size (String.fromInt model.size) "Size"
                ]

        formFields =
            row [] [ leftSide, rightSide ]

        buttons =
            row []
                [ Input.button defaultButtonAttrs
                    { onPress = Just DoSearch
                    , label = text "Query"
                    }
                , Input.button defaultButtonAttrs
                    { onPress = Just ToggleAdvancedSearch
                    , label = text "AdvancedSearch"
                    }
                ]
    in
    row []
        [ column [] [ formFields, buttons ]
        ]


viewRawSearchForm : RawSearchForm -> Element Msg
viewRawSearchForm model =
    column
        []
        [ Input.multiline
            [ height shrink ]
            { onChange = \str -> UpdateRawSearch (Query str)
            , text = model.query
            , placeholder = Nothing
            , label = Input.labelAbove [] (text "Query")
            , spellcheck = False
            }
        , row []
            [ Input.button defaultButtonAttrs
                { onPress = Just DoRawSearch
                , label = text "Raw query"
                }
            , Input.button defaultButtonAttrs
                { onPress = Just ToggleAdvancedSearch
                , label = text "Regular search"
                }
            ]
        ]


viewSearchResults : Int -> SearchStatus -> SearchResults -> String -> Element Msg
viewSearchResults windowWidth status searchResults inboxUrl =
    let
        threadUrl =
            \id ->
                inboxUrl ++ "#inbox/" ++ id

        messageSummaries : List Message -> Element Msg
        messageSummaries messages =
            let
                messageRows : List (Element Msg)
                messageRows =
                    let
                        messageRow : Message -> Element Msg
                        messageRow message =
                            let
                                summary =
                                    row [ spacingXY 20 5, Events.onClick (Toggle message.id) ]
                                        [ el [ width (px 280) ] (text message.date)
                                        , el [ width (px 300), clip ] (text message.from)
                                        , row []
                                            [ el [] <| text message.subject
                                            , el [ Font.color dimmedGray ] <| text (" – " ++ message.snippet)
                                            ]
                                        ]

                                expanded =
                                    let
                                        iframe =
                                            Html.iframe [ Attributes.width windowWidth, Attributes.height graphHeight, Attributes.src ("/message/" ++ message.id) ] []

                                        messageHtml =
                                            let
                                                body =
                                                    Html.Parser.run message.bodyHtml
                                            in
                                            case body of
                                                Ok html ->
                                                    Html.div [] <| Html.Parser.Util.toVirtualDom html

                                                _ ->
                                                    Html.div [] []
                                    in
                                    if message.expanded then
                                        column []
                                            [ link [ Font.color linkColor ]
                                                { url = threadUrl message.threadId
                                                , label = text "Open in Gmail"
                                                }
                                            , el [ width fill ] (html messageHtml)
                                            ]

                                    else
                                        none
                            in
                            column [] [ summary, expanded ]
                    in
                    List.map messageRow messages
            in
            column [] messageRows
    in
    case status of
        Loading ->
            el [] (text "Loading …")

        Success ->
            if List.length searchResults.messages > 0 then
                column []
                    [ el [ width (px graphWidth), height fill ] (html <| Html.div [ Attributes.id "graph" ] [ barGraph (timeSeries searchResults.chartData) ])
                    , messageSummaries searchResults.messages
                    ]

            else
                el [] (text "No messages found")

        Failure e ->
            el [] <| text ("Search error:" ++ e)

        Empty ->
            none


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
        |> required "BodyHtml" string
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
