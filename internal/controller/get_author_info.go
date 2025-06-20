package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorInfo(ctx context.Context, req *library.GetAuthorInfoRequest) (*library.GetAuthorInfoResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := i.authorUseCase.GetAuthor(ctx, req.GetId())

	if err != nil {
		return nil, i.convertError(err)
	}

	return response, nil
}
