package commands

import "ZakuBot/mongo"

func GetUserMoney(userId string) int {
	user, _ := mongo.GetUser(userId)
	return user["money"].(int)
}
