package commands

import "ZakuBot/mongo"

func Register(userID string, userName string) string {
	status := mongo.RegisterUser(userID, userName)
	return status
}
