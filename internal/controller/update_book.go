package controller

import (
	"context"

	"github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) UpdateBook(ctx context.Context, req *library.UpdateBookRequest) (*library.UpdateBookResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := i.booksUseCase.ChangeBookInfo(ctx, req.GetId(), req.GetName(), req.GetAuthorIds())

	if err != nil {
		return nil, i.convertError(err)
	}

	return &library.UpdateBookResponse{}, nil
}
