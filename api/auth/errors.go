package auth

import "github.com/danielgtaylor/huma/v2"

type UnauthorizedGameSessionRequest struct {
}

func (u UnauthorizedGameSessionRequest) ErrorDetail() *huma.ErrorDetail {
	return &huma.ErrorDetail{
		Message:  "not authorized to perform that action",
		Location: "",
		Value:    nil,
	}
}

func (u UnauthorizedGameSessionRequest) Error() string {
	//TODO implement me
	panic("implement me")
}
