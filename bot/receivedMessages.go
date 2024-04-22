package bot

import "ZakuBot/mongo"

func register(userID string) string {
	status := mongo.RegisterUser(userID)
	return status
}
