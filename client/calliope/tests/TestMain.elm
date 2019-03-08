module TestMain exposing (searchResultsDecoderTest, timeSeriesTest, updateTest)

import Expect
import Json.Decode as Decode
import Main exposing (..)
import Test exposing (..)
import Time exposing (millisToPosix)
import Url exposing (Protocol(..))



-- Check out http://package.elm-lang.org/packages/elm-community/elm-test/latest to learn more about testing in Elm!


timeSeriesTest : Test
timeSeriesTest =
    describe "timeSeries"
        [ test "Converting api format to List (Time.Posix, Float)" <|
            \_ ->
                let
                    data =
                        [ { date = "2019-01-01", messages = 1 }
                        , { date = "2019-01-02", messages = 3 }
                        , { date = "2019-01-03", messages = 1 }
                        , { date = "2019-01-04", messages = 2 }
                        , { date = "invalid date", messages = 6 }
                        ]

                    converted =
                        Main.timeSeries data

                    -- Could use Iso8601 to do this instead, but this spells out what's happening.
                    daysFromEpochTo2019 =
                        17897

                    millisInDay =
                        3600 * 1000 * 24

                    jan date =
                        millisToPosix <| (daysFromEpochTo2019 * 3600 * 1000 * 24) + (millisInDay * (date - 1))

                    expected =
                        [ ( jan 1, 1.0 )
                        , ( jan 2, 3.0 )
                        , ( jan 3, 1.0 )
                        , ( jan 4, 2.0 )
                        , ( millisToPosix 0, 6 )
                        ]
                in
                Expect.equal converted expected
        ]


searchResultsDecoderTest : Test
searchResultsDecoderTest =
    let
        result =
            Decode.decodeString searchResultsDecoder jsonData

        expected =
            { chartData = [ { date = "2019-01-01", messages = 1 } ]
            , messages =
                [ annasMessage ]
            , query = "{\n  \"query\": {\n    \"match_all\": {}\n  },\n  \"size\": 10,\n  \"sort\": [\n    {\n      \"Date\": {\n        \"order\": \"desc\"\n      }\n    }\n  ]\n}"
            }
    in
    describe "searchResultsDecoder"
        [ test "decoding API response" <|
            \_ ->
                case result of
                    Ok decoded ->
                        Expect.equal decoded expected

                    Err e ->
                        Expect.fail <| "Problem decoding test data" ++ Decode.errorToString e
        ]


updateTest : Test
updateTest =
    describe "update"
        [ toggleIdTests ]


toggleIdTests =
    describe "Toggle id"
        [ describe "when target is not expanded" <|
            let
                ( updated, _ ) =
                    update (Toggle "2") defaultModel
            in
            [ test "the target should be listed as expanded" <|
                \_ ->
                    Expect.equal "2" updated.expandedMessageId
            , test "the target should have parsed html" <|
                \_ ->
                    let
                        expanded =
                            let
                                appendIfParsed : MessageWrapper -> List String -> List String
                                appendIfParsed ( message, parsed ) parsedList =
                                    if parsed /= Nothing then
                                        message.id :: parsedList

                                    else
                                        parsedList
                            in
                            List.foldl appendIfParsed [] updated.searchResults.messagesWithHtml
                    in
                    Expect.equalLists [ "2" ] expanded
            ]
        , describe "when target is already expanded" <|
            let
                model =
                    { defaultModel | expandedMessageId = "2" }

                ( updated, _ ) =
                    update (Toggle "2") model
            in
            [ test "the target should be collapsed" <|
                \_ ->
                    Expect.equal "" updated.expandedMessageId
            ]
        ]


emptySearchResults =
    SearchResults "" [] []


threeWrappedMessages =
    let
        anna1 =
            { annasMessage | id = "1" }

        anna2 =
            { annasMessage | id = "2" }

        anna3 =
            { annasMessage | id = "3" }
    in
    [ ( anna1, Nothing )
    , ( anna2, Nothing )
    , ( anna3, Nothing )
    ]


searchResults : SearchResults
searchResults =
    { emptySearchResults | messagesWithHtml = threeWrappedMessages }


url =
    { protocol = Http
    , host = "localhost"
    , port_ = Just 3000
    , path = "/"
    , query = Nothing
    , fragment = Nothing
    }


defaultModel : Model
defaultModel =
    { url = url
    , gmailUrl = "https://mail.google.com/mail/"
    , searchForm = SearchForm "" "" "" "" "" "" False "Date" False 100
    , rawSearchForm = RawSearchForm ""
    , searchResults = searchResults
    , expandedMessageId = ""
    , searchStatus = Empty
    , showAdvancedSearch = False
    , windowWidth = 800
    }


annasMessage =
    { body = "Elsa?\nDo you want to build a snowman?\nCome on, let's go and play!\nI never see you anymore\nCome out the door\nIt's like you've gone away\nWe used to be best buddies\nAnd now we're not\nI wish you would tell me why!\nDo you want to build a snowman?\nIt doesn't have to be a snowman\nGo away, Anna\nOkay, bye\nDo you want to build a snowman?\nOr ride our bike around the halls?\nI think some company is overdue\nI've started talking to\nThe pictures on the walls!\nIt gets a little lonely\nAll these empty rooms\nJust watching the hours tick by\n(tick-tock tick-tock tick-tock tick-tock)\nElsa, please I know you're in there\nPeople are asking where you've been\nThey say, \"have courage\" and I'm trying to\nI'm right out here for you\nJust let me in\nWe only have each other\nIt's just you and me\nWhat are we gonna do?\nDo you want to build a snowman?\n"
    , bodyHtml = "<pre>Elsa?\nDo you want to build a snowman?\nCome on, let's go and play!\nI never see you anymore\nCome out the door\nIt's like you've gone away\nWe used to be best buddies\nAnd now we're not\nI wish you would tell me why!\nDo you want to build a snowman?\nIt doesn't have to be a snowman\nGo away, Anna\nOkay, bye\nDo you want to build a snowman?\nOr ride our bike around the halls?\nI think some company is overdue\nI've started talking to\nThe pictures on the walls!\nIt gets a little lonely\nAll these empty rooms\nJust watching the hours tick by\n(tick-tock tick-tock tick-tock tick-tock)\nElsa, please I know you're in there\nPeople are asking where you've been\nThey say, \"have courage\" and I'm trying to\nI'm right out here for you\nJust let me in\nWe only have each other\nIt's just you and me\nWhat are we gonna do?\nDo you want to build a snowman?\n</pre>"
    , cc = ""
    , date = "2019-01-01T00:00:00-08:00"
    , downloadedStartedAt = "2019-01-01T00:00:00-08:00"
    , from = "anna@example.com"
    , id = "id123"
    , labelIds = [ "UNREAD", "CATEGORY_SOCIAL", "INBOX" ]
    , snippet = "Elsa?\nDo you want to build a snowman?\nCome on, let's go and play!\n"
    , subject = "Want to build a snowman?"
    , threadId = "thread123"
    , to = "elsa@example.com"
    }


jsonData =
    """
    {
      "Query": "{\\n  \\"query\\": {\\n    \\"match_all\\": {}\\n  },\\n  \\"size\\": 10,\\n  \\"sort\\": [\\n    {\\n      \\"Date\\": {\\n        \\"order\\": \\"desc\\"\\n      }\\n    }\\n  ]\\n}",
      "ChartData": [
        {
          "Date": "2019-01-01",
          "Messages": 1
        }
      ],

      "Messages":[    {
        "Id": "id123",
        "Url": "something to ignore",
        "ThreadId": "thread123",
        "LabelIds": [
          "UNREAD",
          "CATEGORY_SOCIAL",
          "INBOX"
        ],
        "Date": "2019-01-01T00:00:00-08:00",
        "DownloadedStartedAt": "2019-01-01T00:00:00-08:00",
        "To": "elsa@example.com",
        "Cc": "",
        "From": "anna@example.com",
        "Subject": "Want to build a snowman?",
        "Snippet": "Elsa?\\nDo you want to build a snowman?\\nCome on, let's go and play!\\n",
        "Body": "Elsa?\\nDo you want to build a snowman?\\nCome on, let's go and play!\\nI never see you anymore\\nCome out the door\\nIt's like you've gone away\\nWe used to be best buddies\\nAnd now we're not\\nI wish you would tell me why!\\nDo you want to build a snowman?\\nIt doesn't have to be a snowman\\nGo away, Anna\\nOkay, bye\\nDo you want to build a snowman?\\nOr ride our bike around the halls?\\nI think some company is overdue\\nI've started talking to\\nThe pictures on the walls!\\nIt gets a little lonely\\nAll these empty rooms\\nJust watching the hours tick by\\n(tick-tock tick-tock tick-tock tick-tock)\\nElsa, please I know you're in there\\nPeople are asking where you've been\\nThey say, \\"have courage\\" and I'm trying to\\nI'm right out here for you\\nJust let me in\\nWe only have each other\\nIt's just you and me\\nWhat are we gonna do?\\nDo you want to build a snowman?\\n",
        "BodyHtml": "<pre>Elsa?\\nDo you want to build a snowman?\\nCome on, let's go and play!\\nI never see you anymore\\nCome out the door\\nIt's like you've gone away\\nWe used to be best buddies\\nAnd now we're not\\nI wish you would tell me why!\\nDo you want to build a snowman?\\nIt doesn't have to be a snowman\\nGo away, Anna\\nOkay, bye\\nDo you want to build a snowman?\\nOr ride our bike around the halls?\\nI think some company is overdue\\nI've started talking to\\nThe pictures on the walls!\\nIt gets a little lonely\\nAll these empty rooms\\nJust watching the hours tick by\\n(tick-tock tick-tock tick-tock tick-tock)\\nElsa, please I know you're in there\\nPeople are asking where you've been\\nThey say, \\"have courage\\" and I'm trying to\\nI'm right out here for you\\nJust let me in\\nWe only have each other\\nIt's just you and me\\nWhat are we gonna do?\\nDo you want to build a snowman?\\n</pre>",

        "Source": {}
      }  ]
    }



    """
