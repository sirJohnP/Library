package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetBookInfo(ctx context.Context, req *library.GetBookInfoRequest) (*library.GetBookInfoResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := i.booksUseCase.GetBook(ctx, req.GetId())

	if err != nil {
		return nil, i.convertError(err)
	}

	return response, nil
}
