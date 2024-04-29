package commands

import "ZakuBot/mongo"

func Register(userID string, viewName string, username string) string {
	status := mongo.RegisterUser(userID, viewName, username)
	return status
}

func IsRegistered(userID string) bool {
	_, err := mongo.GetUser(userID)
	if err != nil {
		return false
	}
	return true
}
