package useraggregator

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

var (
	ErrTimeout   = errors.New("timeout_exceeded")
	ErrCancelled = errors.New("context_cancelled")
)

type UserAggregate struct {
	User   string
	Orders int
}

type UserAggregator interface {
	Aggregate(ctx context.Context, id int) (*UserAggregate, error)
}

type userAggregator struct {
	timeout time.Duration
	logger  *slog.Logger

	profileService ProfileService
	orderService   OrderService
}

type Option func(a *userAggregator)

func WithTimeout(t time.Duration) Option {
	return func(a *userAggregator) {
		a.timeout = t
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(a *userAggregator) {
		a.logger = l
	}
}

func NewUserAggregator(
	profileService ProfileService,
	orderService OrderService,
	options ...Option,
) *userAggregator {
	u := &userAggregator{
		profileService: profileService,
		orderService:   orderService,
		timeout:        10 * time.Second,
		logger:         slog.Default(),
	}

	for _, opt := range options {
		opt(u)
	}

	return u
}

func (a *userAggregator) Aggregate(ctx context.Context, id int) (*UserAggregate, error) {
	var name string
	var orders int

	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		a.logger.Info("Fetching user name")
		n, err := a.profileService.GetUserName(ctx, id)
		name = n
		return err
	})

	g.Go(func() error {
		a.logger.Info("Fetching orders count")
		c, err := a.orderService.GetOrdersCount(ctx, id)
		orders = c
		return err
	})

	if err := g.Wait(); err != nil {
		a.logger.Error("Can't fetch user aggregate")
		return nil, err
	}

	return &UserAggregate{
		User:   name,
		Orders: orders,
	}, nil
}
