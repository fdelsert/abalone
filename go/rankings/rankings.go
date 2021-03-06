package rankings

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/ChrisHines/GoSkills/skills"
	"github.com/ChrisHines/GoSkills/skills/trueskill"
	"github.com/danmane/abalone/go/game"
)

type Rating struct {
	Mean   float64
	Stddev float64
}

type Ranking struct {
	PlayerID int64
	Rating   Rating
	Rank     int
}

type Rankings []Ranking

func (r Rankings) Len() int {
	return len(r)
}

func (r Rankings) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Rankings) Less(i, j int) bool {
	if r[i].Rating.Mean == r[j].Rating.Mean {
		// if mean is the same, the better player (i.e. one with the lower rank) will be the one we are more confident about
		return r[i].Rating.Stddev < r[j].Rating.Stddev
	}
	// otherwise, the better player (i.e. one with lower rank) is the one with the higher mean
	return r[i].Rating.Mean > r[j].Rating.Mean
}

func DefaultRatings(players []int64) map[int64]Rating {
	out := make(map[int64]Rating)
	for _, p := range players {
		out[p] = Rating{Mean: gameInfo.InitialMean, Stddev: gameInfo.InitialStddev}
	}
	return out
}

type Result struct {
	WhiteID int64
	BlackID int64
	Outcome game.Outcome
}

func outcomeToRanks(o game.Outcome) ([]int, error) {
	var result []int
	switch o {
	case game.WhiteWins:
		result = []int{1, 2}
	case game.BlackWins:
		result = []int{2, 1}
	case game.Tie:
		result = []int{1, 1}
	default:
		return nil, fmt.Errorf("got null outcome or other invalid outcome: %s", o)
	}
	return result, nil
}

func rating2srating(r Rating) skills.Rating {
	return skills.NewRating(r.Mean, r.Stddev)
}

func srating2rating(s skills.Rating) Rating {
	return Rating{Mean: s.Mean(), Stddev: s.Stddev()}
}

func (r Rating) String() string {
	return fmt.Sprintf("{μ:%.6g σ:%.6g}", r.Mean, r.Stddev)
}

func RateGames(players []int64, games []Result) (Rankings, error) {

	// TODO(dm): this function is messy because the API we are depending on
	// (GoSkills) is pretty weird. Could be cleaned up by just moving the
	// calculation logic into our own impl

	if err := checkGameParticipantsArePlayers(players, games); err != nil {
		return nil, err
	}

	ratings := DefaultRatings(players)
	for _, r := range games {
		whiteTeam := skills.NewTeam()
		whiteTeam.AddPlayer(r.WhiteID, rating2srating(ratings[r.WhiteID]))
		blackTeam := skills.NewTeam()
		blackTeam.AddPlayer(r.BlackID, rating2srating(ratings[r.BlackID]))

		var twoPlayerCalc trueskill.TwoPlayerCalc
		ranks, err := outcomeToRanks(r.Outcome)
		if err != nil {
			return nil, err
		}
		newSkills := twoPlayerCalc.CalcNewRatings(gameInfo, []skills.Team{whiteTeam, blackTeam}, ranks...)
		ratings[r.WhiteID] = srating2rating(newSkills[r.WhiteID])
		ratings[r.BlackID] = srating2rating(newSkills[r.BlackID])
	}

	ranks := make(Rankings, 0)
	for id, rating := range ratings {
		ranks = append(ranks, Ranking{PlayerID: id, Rating: rating, Rank: -1})
	}
	sort.Sort(ranks)
	var prevRating Rating
	for i, _ := range ranks {
		if prevRating == ranks[i].Rating {
			ranks[i].Rank = ranks[i-1].Rank
		} else {
			ranks[i].Rank = i + 1
		}
		prevRating = ranks[i].Rating
	}
	return ranks, nil
}

// ProposeGame returns the IDs of two players who should play the next game.
func (rankings Rankings) ProposeGame() (int64, int64, error) {
	if len(rankings) == 0 {
		return 0, 0, errors.New("cannot propose games when there are no rankings")
	}
	maxUncertainty := math.Inf(-1)
	var id1 int64 = -1
	var idx1 int64 = -1
	for i, rank := range rankings {
		if rank.Rating.Stddev > maxUncertainty {
			maxUncertainty = rank.Rating.Stddev
			id1 = rank.PlayerID
			idx1 = int64(i)
		}
	}
	var id2 int64
	if idx1 == 0 {
		id2 = rankings[1].PlayerID
	} else {
		id2 = rankings[idx1-1].PlayerID
	}
	return id1, id2, nil
}

func checkGameParticipantsArePlayers(players []int64, games []Result) error {
	playerset := make(map[int64]struct{})
	for _, p := range players {
		playerset[p] = struct{}{}
	}
	for _, g := range games {
		ps := []int64{g.WhiteID, g.BlackID}
		for _, p := range ps {
			if _, ok := playerset[p]; !ok {
				return fmt.Errorf("player %d participated in game, but isn't in list of players", p)
			}
		}
	}
	return nil
}

var gameInfo = skills.DefaultGameInfo
