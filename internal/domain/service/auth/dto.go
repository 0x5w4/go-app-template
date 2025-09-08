package auth

import (
	"goapptemp/internal/adapter/util/token"
	"goapptemp/internal/domain/entity"
)

type AuthParams struct {
	Token  string
	Claims *token.Claims
}

type LoginRequest struct {
	User *entity.User
}
