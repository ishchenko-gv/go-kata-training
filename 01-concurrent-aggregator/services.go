package main

import "context"

type ProfileService interface {
	GetUserName(ctx context.Context, id int) (string, error)
}

type OrderService interface {
	GetOrdersCount(ctx context.Context, id int) (int, error)
}

type ProfileServiceMock struct {
	GetUserNameFunc func(ctx context.Context, id int) (string, error)
}

var _ ProfileService = (*ProfileServiceMock)(nil)

func (s *ProfileServiceMock) Reset() {
	*s = ProfileServiceMock{}
}

func (s *ProfileServiceMock) GetUserName(ctx context.Context, id int) (string, error) {
	return s.GetUserNameFunc(ctx, id)
}

type OrderServiceMock struct {
	GetOrdersCountFunc func(ctx context.Context, id int) (int, error)
}

var _ OrderService = (*OrderServiceMock)(nil)

func (s *OrderServiceMock) Reset() {
	*s = OrderServiceMock{}
}

func (s *OrderServiceMock) GetOrdersCount(ctx context.Context, id int) (int, error) {
	return s.GetOrdersCountFunc(ctx, id)
}
