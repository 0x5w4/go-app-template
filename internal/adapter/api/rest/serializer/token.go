package serializer

import (
	"goapptemp/internal/domain/entity"
	"time"
)

type TokenResponseData struct {
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresAt  string `json:"access_token_expires_at"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresAt string `json:"refresh_token_expires_at"`
	TokenType             string `json:"token_type"`
}

func SerializeToken(arg *entity.Token) *TokenResponseData {
	if arg == nil {
		return nil
	}

	return &TokenResponseData{
		AccessToken:           arg.AccessToken,
		AccessTokenExpiresAt:  arg.AccessTokenExpiresAt.Format(time.RFC3339),
		RefreshToken:          arg.RefreshToken,
		RefreshTokenExpiresAt: arg.RefreshTokenExpiresAt.Format(time.RFC3339),
		TokenType:             arg.TokenType,
	}
}

func SerializeTokens(arg []*entity.Token) []*TokenResponseData {
	if len(arg) == 0 {
		return nil
	}

	res := make([]*TokenResponseData, 0, len(arg))

	for i := range arg {
		if arg[i] == nil {
			continue
		}

		res = append(res, SerializeToken(arg[i]))
	}

	return res
}
