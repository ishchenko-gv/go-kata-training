package main_test

import (
	"context"
	main "conurrent-aggregator"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAggregator(t *testing.T) {
	ctx := context.Background()
	profileServiceMock := &main.ProfileServiceMock{}
	orderServiceMock := &main.OrderServiceMock{}
	resetMocks := func() {
		profileServiceMock.Reset()
		orderServiceMock.Reset()
	}

	userAggregator := main.NewUserAggregator(
		profileServiceMock,
		orderServiceMock,
	)

	t.Run("success", func(t *testing.T) {
		defer resetMocks()
		profileServiceMock.GetUserNameFunc = func(ctx context.Context, id int) (string, error) {
			return "Alice", nil
		}

		orderServiceMock.GetOrdersCountFunc = func(ctx context.Context, id int) (int, error) {
			return 5, nil
		}

		got, err := userAggregator.Aggregate(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, &main.UserAggregate{
			User:   "Alice",
			Orders: 5,
		}, got)
	})
}
