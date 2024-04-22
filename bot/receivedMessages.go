package bot

import (
	"ZakuBot/mongo"
)

func register(userID string) string {
	status := mongo.RegisterUser(userID)
	return status
}

func drawCards() (string, error) {
	cards, err := mongo.DrawCards()
	if err != nil {
		return "", err
	}
	return CombineDrawnCards(cards), nil
}
