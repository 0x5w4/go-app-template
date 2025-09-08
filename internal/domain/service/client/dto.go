package client

import (
	"goapptemp/internal/adapter/repository/mysql"
	"goapptemp/internal/domain/entity"
	"goapptemp/internal/domain/service/auth"
)

type CreateClientRequest struct {
	AuthParams *auth.AuthParams
	Client     *entity.Client
}

type UpdateClientRequest struct {
	AuthParams *auth.AuthParams
	Update     *mysql.UpdateClientPayload
}

type DeleteClientRequest struct {
	AuthParams *auth.AuthParams
	ClientID   uint
}

type FindClientsRequest struct {
	AuthParams *auth.AuthParams
	Filter     *mysql.FilterClientPayload
}

type FindOneClientRequest struct {
	AuthParams *auth.AuthParams
	ClientID   uint
}

type IsDeletableClientRequest struct {
	AuthParams *auth.AuthParams
	ClientID   uint
}
