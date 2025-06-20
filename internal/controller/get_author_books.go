package controller

import (
	generated "github.com/project/library/generated/api/library"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *implementation) GetAuthorBooks(req *generated.GetAuthorBooksRequest, server generated.Library_GetAuthorBooksServer) error {
	if err := req.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	books, err := i.booksUseCase.GetBooksByAuthor(server.Context(), req.GetAuthorId())

	if err != nil {
		return i.convertError(err)
	}

	for _, book := range books {
		if err := server.Send(book); err != nil {
			return i.convertError(err)
		}
	}

	return nil
}
