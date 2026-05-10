package main_test

import (
	"context"
	main "conurrent-aggregator"
	"errors"
	"testing"
	"time"

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

	t.Run("success", func(t *testing.T) {
		userAggregator := main.NewUserAggregator(
			profileServiceMock,
			orderServiceMock,
		)

		got, err := userAggregator.Aggregate(ctx, 1)
		assert.NoError(t, err)

		assert.Equal(t, &main.UserAggregate{
			User:   "Alice",
			Orders: 5,
		}, got)
	})

	t.Run("timeout", func(t *testing.T) {
		defer resetMocks()
		orderServiceMock.GetOrdersCountFunc = func(ctx context.Context, id int) (int, error) {
			select {
			case <-ctx.Done():
				return 0, main.ErrTimeout
			case <-time.After(5 * time.Second):
				return 5, nil
			}
		}

		userAggregator := main.NewUserAggregator(
			profileServiceMock,
			orderServiceMock,
			main.WithTimeout(3*time.Second),
		)

		got, err := userAggregator.Aggregate(ctx, 1)
		assert.ErrorIs(t, err, main.ErrTimeout)
		assert.Nil(t, got)
	})

	t.Run("fail_early", func(t *testing.T) {
		defer resetMocks()
		profileServiceMock.GetUserNameFunc = func(ctx context.Context, id int) (string, error) {
			return "", errors.New("failed to fetch user name")
		}

		var ordersCountCalled bool
		orderServiceMock.GetOrdersCountFunc = func(ctx context.Context, id int) (int, error) {
			var err error
			select {
			case <-ctx.Done():
				err = main.ErrCancelled
			case <-time.After(3 * time.Second):
			}
			assert.ErrorIs(t, err, main.ErrCancelled)
			ordersCountCalled = true
			return 0, err
		}

		userAggregator := main.NewUserAggregator(
			profileServiceMock,
			orderServiceMock,
		)

		got, err := userAggregator.Aggregate(ctx, 1)
		assert.Error(t, err)

		assert.Nil(t, got)
		assert.True(t, ordersCountCalled)
	})
}
