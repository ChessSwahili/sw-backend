package data

import (
	"database/sql"
)

type Models struct {
	Users  UserModel
	Tokens TokenModel
	Accounts AccountModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
		Accounts: AccountModel{DB: db},
	}
}
