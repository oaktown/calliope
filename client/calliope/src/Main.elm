module Main exposing (ChartDay, Message, Model, Msg(..), RawSearchForm, RawSearchFormMsg(..), SearchForm, SearchFormMsg(..), SearchResults, formCheckBox, formField, formInput, init, main, searchResultsDecoder, subscriptions, timeSeries, update, updateRawSearchForm, updateSearchForm, view, viewRawSearchForm, viewSearchForm, viewSearchForms)

import BarGraph exposing (barGraph)
import Browser
import Debug
import Html exposing (Html, a, br, button, div, h1, h2, input, label, li, p, table, tbody, td, text, textarea, th, thead, tr, ul)
import Html.Attributes exposing (checked, class, cols, href, id, name, placeholder, rows, scope, type_, value)
import Html.Events exposing (onClick, onInput)
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder, field, int, list, string)
import Json.Decode.Pipeline exposing (required)
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
    , searchResults : Result Http.Error SearchResults
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
    }


type alias ChartDay =
    { date : String
    , messages : Int
    }


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
            Ok (SearchResults "" [] [])
    in
    ( { gmailUrl = "https://mail.google.com/mail/"
      , searchForm = defaultSearchForm
      , rawSearchForm = defaultRawSearchForm
      , searchResults = emptySearchResults
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


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UpdateGmailUrl gmailUrl ->
            ( { model | gmailUrl = gmailUrl }, Cmd.none )

        UpdateRawSearch field ->
            ( { model | rawSearchForm = updateRawSearchForm field model.rawSearchForm }, Cmd.none )

        UpdateSearch field ->
            ( { model | searchForm = updateSearchForm field model.searchForm }, Cmd.none )

        DoSearch ->
            ( model, doSearch model.searchForm )

        DoRawSearch ->
            ( model, doRawSearch model.rawSearchForm )

        GotSearch searchResults ->
            ( { model | searchResults = searchResults }, Cmd.none )


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
                    model



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    div []
        [ div [] [ viewSearchForms model ]
        , div [] [ viewSearchResults model.searchResults model.gmailUrl ]
        ]


viewSearchForms : Model -> Html Msg
viewSearchForms model =
    div [ class "search-form" ]
        [ formInput "gmailurl" "gmailurl" model.gmailUrl UpdateGmailUrl
        , viewSearchForm model.searchForm
        , div [] [ p [] [ text "OR" ] ]
        , viewRawSearchForm model.rawSearchForm
        ]


viewSearchForm : SearchForm -> Html Msg
viewSearchForm model =
    let
        searchFormMessage : (a -> SearchFormMsg) -> (a -> Msg)
        searchFormMessage fn =
            \a -> UpdateSearch (fn a)
    in
    div []
        [ formInput "participants" "Participants (applies to From:, To:, and CC:):" model.participants (searchFormMessage Participants)
        , formInput "bodyOrSubject" "Body or subject:" model.bodyOrSubject (searchFormMessage BodyOrSubject)
        , formInput "startDate" "Start date (\"YYYY-MM-DD\"):" model.startDate (searchFormMessage StartDate)
        , formInput "endDate" "End date (\"YYYY-MM-DD\"):" model.endDate (searchFormMessage EndDate)
        , formInput "timezone" "Time zone (e.g. -0800 for PST):" model.timeZone (searchFormMessage TimeZone)
        , formInput "label" "Label:" model.label (searchFormMessage Label)
        , formCheckBox "starred" "Starred only:" model.starredOnly (UpdateSearch StarredOnly)
        , formInput "sort" "Sort field:" model.sortField (searchFormMessage SortField)
        , formCheckBox "ascending" "Ascending:" model.ascending (UpdateSearch Ascending)
        , formInput "size" "Size:" (String.fromInt model.size) (searchFormMessage Size)
        , btn "query" DoSearch
        ]


viewRawSearchForm : RawSearchForm -> Html Msg
viewRawSearchForm model =
    let
        rawSearchFormMessage : (String -> RawSearchFormMsg) -> (String -> Msg)
        rawSearchFormMessage fn =
            \str -> UpdateRawSearch (fn str)

        queryField =
            textarea [ rows 12, cols 120, name "query", value model.query, onInput (rawSearchFormMessage Query) ] []
    in
    div []
        [ formField "Query:" queryField
        , btn "raw query" DoRawSearch
        ]


formInput : String -> String -> String -> (String -> msg) -> Html msg
formInput fieldName labelText val msgFn =
    let
        f =
            input [ name fieldName, placeholder fieldName, value val, onInput msgFn ] []
    in
    formField labelText f


formCheckBox : String -> String -> Bool -> msg -> Html msg
formCheckBox fieldName labelText val msg =
    let
        f =
            input [ type_ "checkbox", name fieldName, checked val, onClick msg ] []
    in
    formField labelText f


formField : String -> Html msg -> Html msg
formField labelText f =
    div [] [ label [] [ text labelText, br [] [], f ] ]


btn : String -> Msg -> Html Msg
btn label msg =
    button [ onClick msg ] [ text label ]


viewSearchResults : Result Http.Error SearchResults -> String -> Html Msg
viewSearchResults results inboxUrl =
    let
        threadUrl =
            \id ->
                inboxUrl ++ "#inbox/" ++ id

        summaryRow =
            \message ->
                tr []
                    [ td [] [ text message.date ]
                    , td [] [ text message.from ]
                    , td [] [ a [ href (threadUrl message.threadId) ] [ text "Gmail" ] ]
                    , td [] [ a [ href ("#" ++ message.id) ] [ text "Jump" ] ]
                    , td [] [ text message.subject ]
                    ]

        toc =
            \messages ->
                div []
                    [ table []
                        [ thead []
                            [ tr []
                                [ th [] [ text "Date" ]
                                , th [] [ text "From" ]
                                , th [] [ text "Gmail" ]
                                , th [] [ text "Jump" ]
                                , th [] [ text "Subject" ]
                                ]
                            ]
                        , tbody []
                            (List.map summaryRow messages)
                        ]
                    ]

        messagesView =
            \messages ->
                div [] (List.map messageView messages)

        messageView =
            \message ->
                let
                    row =
                        \label value ->
                            tr []
                                [ th [ scope "row" ] [ text label ]
                                , td [] [ text value ]
                                ]
                in
                li []
                    [ table [ id message.id ]
                        [ row "Date" message.date
                        , row "From" message.from
                        , row "To" message.to
                        , row "Cc" message.cc
                        , row "Subject" message.subject
                        ]
                    , div [] [ text message.body ]
                    , ul []
                        [ li []
                            [ a [ href "#ToC" ] [ text "ToC" ]
                            ]
                        , li []
                            [ a [ href (threadUrl message.threadId) ] [ text "Gmail" ] ]
                        ]
                    ]
    in
    case results of
        Ok searchResults ->
            if List.length searchResults.messages > 0 then
                div []
                    [ div [] [ barGraph (timeSeries searchResults.chartData) ]
                    , h1 [ id "ToC" ] [ text "Table of Contents" ]
                    , toc searchResults.messages
                    , h2 [] [ text "Email messages" ]
                    , messagesView searchResults.messages
                    ]

            else
                div [] []

        Err e ->
            div [] [ text (Debug.toString e) ]


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
            Url.Builder.absolute [ "api", "search" ]
                [ string "query" rawSearchForm.query
                ]
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
