package game

type Account struct {
	Username          string
	Session           *any
	GameServerAddress *string
}

func NewAccount(username string) *Account {
	return &Account{
		Username: username,
	}
}
