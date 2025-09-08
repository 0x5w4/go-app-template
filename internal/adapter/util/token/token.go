package token

type Token interface {
	Generate(userID uint) (string, error)
	Verify(tokenStr string) (*Claims, error)
}
