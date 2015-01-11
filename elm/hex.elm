module Hex where
import List
import Dict
import Misc
type Direction = TopRight | MidRight | BotRight | TopLeft | MidLeft | BotLeft

type alias Position = (Int, Int)

dist2 : Position -> Position -> Int -- distance * 2 (to avoid fractional types)
dist2 (q1,r1) (q2,r2) = abs (q1 - q2) + abs (r1 - r2) + abs (q1 + r1 - q2 - r2)

adjacent : Direction -> Position -> Position
adjacent d (q, r) = case d of 
    TopRight -> (q+1, r-1)
    MidRight -> (q+1, r  )
    BotRight -> (q  , r+1)
    BotLeft  -> (q-1, r+1)
    MidLeft  -> (q-1, r  )
    TopLeft  -> (q  , r-1)

opposite : Direction -> Direction
opposite d = case d of 
    TopRight -> BotLeft
    MidRight -> MidLeft
    BotRight -> TopLeft
    BotLeft  -> TopRight
    MidLeft  -> MidRight
    TopLeft  -> BotRight

directions = [TopRight, MidRight, BotRight, BotLeft, MidLeft, TopLeft]

nearbyDirections : Direction -> List Direction
nearbyDirections d = 
    let idx = Misc.fromJust <| Misc.index d directions
        nearby = map (\x -> (x + idx) % 6) [-1, 0, 1]
    in  List.map (Misc.fromJust << (flip Misc.retrieve <| directions)) nearby

colinear : Direction -> Direction -> Bool
colinear d1 d2 = d1 == d2 || d1 == opposite d2

hexagonalGrid : Int -> List Position
hexagonalGrid rad = List.concat <| List.map ring [0..rad - 1]


tailOrOnly : List a -> List a
tailOrOnly (x::xs) = if xs == [] then [x] else xs

ring : Int -> List Position
ring n = let ds = List.concat <| List.map (List.repeat n << adjacent) directions 
         in tailOrOnly <| List.scanl identity (-n, 0) ds 

findDirection : Position -> Position -> Maybe Direction
findDirection (q1, r1) (q2, r2) = 
    if
        | q1 == q2 && r1 == r2 -> Nothing
        | q1 == q2             -> if r1 < r2 then Just BotRight else Just TopLeft
        | r1 == r2             -> if q1 < q2 then Just MidRight else Just MidLeft
        | r1 + q1 == r2 + q2   -> if q1 < q2 then Just TopRight else Just BotLeft
        | otherwise            -> Nothing