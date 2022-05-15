package repository

import "github.com/ohmpatel1997/vwap/entity"

//go:generate mockery --name Repository --case underscore --output ../../pkg/mocks/repository --outpkg repository
//go:generate mockery --name MatchRepository --case underscore --output ../../pkg/mocks/repository --outpkg repository

// Repository – interface for all repositories
type Repository interface {
	Match() MatchRepository
}

type repository struct {
	match MatchRepository
}

// NewRepository – constructor for repositories
func NewRepository(config *entity.Config) Repository {
	return &repository{
		match: newMatchRepository(config),
	}
}

func (m *repository) Match() MatchRepository {
	return m.match
}
