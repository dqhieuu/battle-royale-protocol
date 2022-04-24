package game

type Account struct {
	Username          string
	GameServerAddress *string
}

func NewAccount(username string) *Account {
	return &Account{
		Username: username,
	}
}
