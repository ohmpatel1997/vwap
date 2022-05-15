package usecase

import (
	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/producers"
	"github.com/ohmpatel1997/vwap/repository"
)

//go:generate mockery --name UseCase --case underscore --output ../../pkg/mocks/usecase --outpkg usecase
//go:generate mockery --name MatchUseCase --case underscore --output ../../pkg/mocks/usecase --outpkg usecase

// UseCase – interface for all useCases
type UseCase interface {
	Match() MatchUseCase
}

type useCase struct {
	match MatchUseCase
}

// NewUseCase – constructor for creating useCase manager
func NewUseCase(
	repo repository.Repository,
	producer producers.Producer,
	config *entity.Config,
) UseCase {
	return &useCase{
		match: newMatchUseCase(repo, producer, config),
	}
}

func (m *useCase) Match() MatchUseCase {
	return m.match
}
