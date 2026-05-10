package main

import (
	"context"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

type UserAggregate struct {
	User   string
	Orders int
}

type UserAggregator interface {
	Aggregate(ctx context.Context, id int) (*UserAggregate, error)
}

type userAggregator struct {
	// TODO: add functional options
	timeout time.Duration
	logger  log.Logger

	profileService ProfileService
	orderService   OrderService
}

func NewUserAggregator(
	profileService ProfileService,
	orderService OrderService,
) *userAggregator {
	return &userAggregator{
		profileService: profileService,
		orderService:   orderService,
	}
}

func (a *userAggregator) Aggregate(ctx context.Context, id int) (*UserAggregate, error) {
	var name string
	var orders int

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		n, err := a.profileService.GetUserName(ctx, id)
		name = n
		return err
	})

	g.Go(func() error {
		c, err := a.orderService.GetOrdersCount(ctx, id)
		orders = c
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &UserAggregate{
		User:   name,
		Orders: orders,
	}, nil
}
