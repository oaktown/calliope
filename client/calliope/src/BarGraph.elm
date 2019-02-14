module BarGraph exposing (barGraph)

-- Adapted from elm-visualization examples

import Axis
import DateFormat
import Scale exposing (BandConfig, BandScale, ContinuousScale, defaultBandConfig)
import Time
import TypedSvg exposing (g, rect, style, svg, text_)
import TypedSvg.Attributes exposing (class, textAnchor, transform, viewBox)
import TypedSvg.Attributes.InPx exposing (height, width, x, y)
import TypedSvg.Core exposing (Svg, text)
import TypedSvg.Types exposing (AnchorAlignment(..), Transform(..))


barGraph : List ( Time.Posix, Float ) -> Svg msg
barGraph timeSeries =
    let
        maxY =
            let
                max t m =
                    if Tuple.second t > m then
                        Tuple.second t

                    else
                        m
            in
            List.foldl max 0 timeSeries

        w : Float
        w =
            900

        h : Float
        h =
            450

        padding : Float
        padding =
            30

        xScale : List ( Time.Posix, Float ) -> BandScale Time.Posix
        xScale model =
            List.map Tuple.first model
                |> Scale.band { defaultBandConfig | paddingInner = 0.1, paddingOuter = 0.2 } ( 0, w - 2 * padding )

        yScale : ContinuousScale Float
        yScale =
            Scale.linear ( h - 2 * padding, 0 ) ( 0, maxY )

        dateFormat : Time.Posix -> String
        dateFormat =
            DateFormat.format [ DateFormat.dayOfMonthFixed, DateFormat.text " ", DateFormat.monthNameAbbreviated ] Time.utc

        xAxis : List ( Time.Posix, Float ) -> Svg msg
        xAxis model =
            Axis.bottom [] (Scale.toRenderable dateFormat (xScale model))

        yAxis : Svg msg
        yAxis =
            Axis.left [ Axis.tickCount 5 ] yScale

        column : BandScale Time.Posix -> ( Time.Posix, Float ) -> Svg msg
        column scale ( date, value ) =
            g [ class [ "column" ] ]
                [ rect
                    [ x <| Scale.convert scale date
                    , y <| Scale.convert yScale value
                    , width <| Scale.bandwidth scale
                    , height <| h - Scale.convert yScale value - 2 * padding
                    ]
                    []
                , text_
                    [ x <| Scale.convert (Scale.toRenderable dateFormat scale) date
                    , y <| Scale.convert yScale value - 5
                    , textAnchor AnchorMiddle
                    ]
                    [ text <| String.fromFloat value ]
                ]

        view : List ( Time.Posix, Float ) -> Svg msg
        view model =
            svg [ viewBox 0 0 w h ]
                [ style [] [ text """
                    .column rect { fill: rgba(0, 50, 255, 0.7); }
                    .column text { display: none; }
                    .column:hover rect { fill: rgb(0, 50, 255); }
                    .column:hover text { display: inline; }
                  """ ]
                , g [ transform [ Translate (padding - 1) (h - padding) ] ]
                    [ xAxis model ]
                , g [ transform [ Translate (padding - 1) padding ] ]
                    [ yAxis ]
                , g [ transform [ Translate padding padding ], class [ "series" ] ] <|
                    List.map (column (xScale model)) model
                ]
    in
    view timeSeries
