package usecase

import (
	"github.com/ohmpatel1997/vwap/entity"
	"github.com/ohmpatel1997/vwap/producers"
	"github.com/ohmpatel1997/vwap/repository"
)

//go:generate mockery --name UseCase --case underscore --output ../../pkg/mocks/usecase --outpkg usecase
type UseCase interface {
	MatchVWAP() MatchUseCase
}

type useCase struct {
	match MatchUseCase
}

func NewUseCase(
	repo repository.Repository,
	producer producers.Producer,
	config *entity.Config,
) UseCase {
	return &useCase{
		match: newMatchUseCase(repo, producer, config),
	}
}

func (m *useCase) MatchVWAP() MatchUseCase {
	return m.match
}
