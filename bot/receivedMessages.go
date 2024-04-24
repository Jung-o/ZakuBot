package bot

import (
	"ZakuBot/mongo"
	"image/jpeg"
	"os"
)

func register(userID string) string {
	status := mongo.RegisterUser(userID)
	return status
}

func combinedCardsFile() (string, error) {
	cards, err := mongo.DrawCards()
	if err != nil {
		return "", err
	}
	// Create a new file
	filepath := "files/combined.jpg"
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	combinedImg := CombineDrawnCards(cards)
	// Encode (write) the image to the file
	if err := jpeg.Encode(file, combinedImg, nil); err != nil {
		panic(err)
	}
	return filepath, nil
}
