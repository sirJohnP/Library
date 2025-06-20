package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) RegisterAuthor(ctx context.Context, req *library.RegisterAuthorRequest) (*library.RegisterAuthorResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := i.authorUseCase.RegisterAuthor(ctx, req.GetName())

	if err != nil {
		return nil, i.convertError(err)
	}

	return response, nil
}
