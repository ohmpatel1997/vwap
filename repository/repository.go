package repository

import "github.com/ohmpatel1997/vwap/entity"

//go:generate mockery --name Repository --case underscore --output ../../pkg/mocks/repository --outpkg repository
type Repository interface {
	Match() MatchRepository
}

type repository struct {
	match MatchRepository
}

func NewRepository(config *entity.Config) Repository {
	return &repository{
		match: newMatchRepository(config),
	}
}

func (m *repository) Match() MatchRepository {
	return m.match
}
