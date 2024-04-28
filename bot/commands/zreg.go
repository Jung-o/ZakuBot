package commands

import "ZakuBot/mongo"

func Register(userID string, userName string) string {
	status := mongo.RegisterUser(userID, userName)
	return status
}

func IsRegistered(userID string) bool {
	_, err := mongo.GetUser(userID)
	if err != nil {
		return false
	}
	return true
}
