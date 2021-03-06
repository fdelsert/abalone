package db

import api "github.com/danmane/abalone/go/api"

type gamesDB struct {
	*resources
}

func (s *gamesDB) List() ([]api.Game, error) {
	var games []api.Game
	if err := s.DB.Find(&games).Error; err != nil {
		return nil, err
	}
	return games, nil
}

func (s *gamesDB) ListDetailed() ([]*api.GameWithDetails, error) {
	var games []*api.GameWithDetails
	if err := s.DB.Find(&games).Error; err != nil {
		return nil, err
	}
	var ids []int64
	for _, g := range games {
		ids = append(ids, g.WhiteId)
		ids = append(ids, g.BlackId)
	}
	var players []api.Player
	if err := s.DB.Where(ids).Find(&players).Error; err != nil {
		return nil, err
	}
	playermap := make(map[int64]*api.Player)
	for _, p := range append(players) {
		playermap[p.ID] = &p
	}
	for _, g := range games {
		g.WhitePlayer = *playermap[g.WhiteId]
		g.BlackPlayer = *playermap[g.BlackId]
	}
	return games, nil
}

var _ api.GamesService = new(gamesDB)
