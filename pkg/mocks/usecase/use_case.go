// Code generated by mockery v2.10.4. DO NOT EDIT.

package usecase

import (
	"github.com/ohmpatel1997/vwap/factory"
	mock "github.com/stretchr/testify/mock"
)

// UseCase is an autogenerated mock type for the UseCase type
type UseCase struct {
	mock.Mock
}

// Match provides a mock function with given fields:
func (_m *UseCase) MatchVWAP() factory.MatchUseCase {
	ret := _m.Called()

	var r0 factory.MatchUseCase
	if rf, ok := ret.Get(0).(func() factory.MatchUseCase); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(factory.MatchUseCase)
		}
	}

	return r0
}