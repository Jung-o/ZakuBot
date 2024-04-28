package commands

import (
	"ZakuBot/mongo"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var vignetteImagePaths = map[color.Color]string{
	color.RGBA{R: 113, G: 179, B: 255, A: 255}: "files/banners/blue.png",
	color.RGBA{R: 78, G: 199, B: 75, A: 255}:   "files/banners/green.png",
	color.RGBA{R: 255, G: 169, B: 59, A: 255}:  "files/banners/orange.png",
	color.RGBA{R: 240, G: 165, B: 216, A: 255}: "files/banners/pink.png",
	color.RGBA{R: 120, G: 53, B: 165, A: 255}:  "files/banners/purple.png",
	color.RGBA{R: 213, G: 58, B: 58, A: 255}:   "files/banners/red.png",
	color.RGBA{R: 255, G: 255, B: 255, A: 255}: "files/banners/white.png",
	color.RGBA{R: 254, G: 255, B: 137, A: 255}: "files/banners/yellow.png",
}

type DropWin struct {
	winnerID      string
	characterID   string
	characterName string
	opponents     int
}

func CombinedCardsFile(cards []bson.M) (string, error) {
	// Create a new file
	filepath := "files/combined.jpg"
	file, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	combinedImg := combineDrawnCards(cards)
	// Encode (write) the image to the file
	if err := jpeg.Encode(file, combinedImg, nil); err != nil {
		panic(err)
	}
	return filepath, nil
}

// combineDrawnCards downloads the images of the drawn cards, adds a vignette based on the average color of the image,
// writes the character's name and series on the image and combines them into a single image
func combineDrawnCards(imagesInfo []bson.M) image.Image {
	images := make([]image.Image, len(imagesInfo))
	for i, card := range imagesInfo {
		characterId := card["_id"].(string)
		imageName := characterId + "-1"
		// Construct the path to the image file
		imagePath := fmt.Sprintf("files/cards/%s.jpg", imageName)
		artworkInfos, err := mongo.GetArtworkInfos(characterId, 1)
		artworkUrl := artworkInfos["artworkUrl"].(string)
		_, err = os.Stat(imagePath)
		if os.IsNotExist(err) {
			// File does not exist, run processImage
			imgRGBA := processImage(characterId, 1, artworkUrl)
			images[i] = imgRGBA
		} else if err != nil {
			// Another error occurred
			log.Fatal(err)
		} else {
			// File exists, open it
			imgRGBA, err := openImage(imagePath)
			if err != nil {
				log.Fatal(err)
			}
			images[i] = imgRGBA
		}
	}

	return combineImages(images)
}

func addLabel(img *image.RGBA, name string, series string) {
	fontBytes, err := os.ReadFile("files/LT_Comical.ttf")
	if err != nil {
		log.Fatal(err)
	}

	f, err := sfnt.Parse(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	if len(name) > 18 {
		name = name[:18] + "-"
	}
	if len(series) > 20 {
		series = series[:20] + "-"
	}

	// Create two font.Face (name and series) from the parsed TTF file
	nameFont, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    25,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	nameDrawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 255}), // White text
		Face: nameFont,
		Dot:  fixed.Point26_6{X: fixed.I(10), Y: fixed.I(img.Bounds().Dy() - 30)}, // Position of the text
	}
	nameDrawer.DrawString(name)

	seriesFont, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	seriesDrawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 255}), // White text
		Face: seriesFont,
		Dot:  fixed.Point26_6{X: fixed.I(10), Y: fixed.I(img.Bounds().Dy() - 5)}, // Position of the text
	}
	seriesDrawer.DrawString(series)
}

// Adds a vignette to the image based on the average color of the image
func addVignetteBasedOnAverageColor(characterImage image.Image) (image.Image, error) {

	// Calculate the average color of the character's image
	var rTotal, gTotal, bTotal, count int64
	for y := characterImage.Bounds().Min.Y; y < characterImage.Bounds().Max.Y; y++ {
		for x := characterImage.Bounds().Min.X; x < characterImage.Bounds().Max.X; x++ {
			r, g, b, _ := characterImage.At(x, y).RGBA()
			rTotal += int64(r)
			gTotal += int64(g)
			bTotal += int64(b)
			count++
		}
	}

	averageColor := color.RGBA{
		R: uint8(rTotal / count),
		G: uint8(gTotal / count),
		B: uint8(bTotal / count),
		A: 255,
	}

	// Select a vignette image based on the closest average color
	vignetteImagePath := closestColor(vignetteImagePaths, averageColor)

	// Load the vignette image
	vignetteImageFile, err := os.Open(vignetteImagePath)
	if err != nil {
		return nil, err
	}
	defer vignetteImageFile.Close()

	vignetteImage, _, err := image.Decode(vignetteImageFile)
	if err != nil {
		return nil, err
	}

	// Resize the vignette image to match the size of the character's image
	vignetteImage = resize.Resize(uint(characterImage.Bounds().Dx()), uint(characterImage.Bounds().Dy()), vignetteImage, resize.Lanczos3)

	// Overlay the vignette image on top of the character's image
	overlay := image.NewRGBA(characterImage.Bounds())
	draw.Draw(overlay, overlay.Bounds(), characterImage, image.Point{}, draw.Src)
	draw.Draw(overlay, overlay.Bounds(), vignetteImage, image.Point{}, draw.Over)

	return overlay, nil
}

// Calculates the closest color to the target color
func closestColor(colors map[color.Color]string, target color.Color) string {
	var (
		closestColor     color.Color
		smallestDistance = math.MaxFloat64
	)

	for c := range colors {
		r, g, b, _ := c.RGBA()
		tr, tg, tb, _ := target.RGBA()

		distance := math.Sqrt(math.Pow(float64(r)-float64(tr), 2) + math.Pow(float64(g)-float64(tg), 2) + math.Pow(float64(b)-float64(tb), 2))

		if distance < smallestDistance {
			smallestDistance = distance
			closestColor = c
		}
	}

	return colors[closestColor]
}

func downloadMissingImage(filePath string, artworkUrl string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Download the images
		resp, err := http.Get(artworkUrl)
		if err != nil {
			fmt.Println("Error downloading image")
			fmt.Println(artworkUrl)
			fmt.Println(filePath)
			log.Fatal(err)
		}
		defer resp.Body.Close()

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			fmt.Println("Error decoding image")
			fmt.Println(artworkUrl)
			fmt.Println(filePath)
			log.Fatal(err)
		}

		// Save the image to the file
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		err = jpeg.Encode(file, img, nil)
		if err != nil {
			fmt.Println("Error encoding image")
			fmt.Println(artworkUrl)
			fmt.Println(filePath)
			log.Fatal(err)
		}
	}
}

func openImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func processImage(characterId string, artworkId int, artworkUrl string) *image.RGBA {
	// Construct the path to the image file
	imageName := fmt.Sprintf("%s-%d", characterId, artworkId)
	imagePath := fmt.Sprintf("files/cards/%s.jpg", imageName)
	downloadMissingImage(imagePath, artworkUrl)
	img, err := openImage(imagePath)

	vignettedImage, err := addVignetteBasedOnAverageColor(img)
	if err != nil {
		log.Fatal(err)
	}

	// Convert image to RGBA
	imgRGBA := image.NewRGBA(vignettedImage.Bounds())
	draw.Draw(imgRGBA, imgRGBA.Bounds(), vignettedImage, image.Point{}, draw.Src)

	// Add character's name and series to the image
	characterInfos, err := mongo.GetCharInfos(characterId)
	if err != nil {
		log.Fatal(err)
	}

	name := characterInfos["name"].(string)
	series := characterInfos["series"].(string)
	addLabel(imgRGBA, name, series)

	// Save the image to the file
	file, err := os.Create(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = jpeg.Encode(file, imgRGBA, nil)
	if err != nil {
		log.Fatal(err)
	}
	return imgRGBA
}

func combineImages(images []image.Image) image.Image {
	width := 0
	height := 0
	for _, img := range images {
		width += img.Bounds().Dx()
		if h := img.Bounds().Dy(); h > height {
			height = h
		}
	}
	combined := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for _, img := range images {
		draw.Draw(combined, image.Rect(x, 0, x+img.Bounds().Dx(), height), img, image.Point{}, draw.Src)
		x += img.Bounds().Dx()
	}

	return combined
}

func MessageCleanup(session *discordgo.Session, message *discordgo.Message) {
	session.MessageReactionsRemoveAll(message.ChannelID, message.ID)
	session.ChannelMessageEdit(message.ChannelID, message.ID,
		fmt.Sprintf("These cards can't be dropped anymore."))
}

func getUsersWhoReacted(session *discordgo.Session, channelID string, messageID string, emoji string) ([]string, error) {
	var users []string

	lastID := ""

	for {
		partialUsers, err := session.MessageReactions(channelID, messageID, emoji, 100, "", lastID)
		if err != nil {
			return nil, err
		}

		if len(partialUsers) == 0 {
			break
		}

		lastID = partialUsers[len(partialUsers)-1].ID
		users = append(users, lastID)
	}

	return users, nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func ChooseWinners(session *discordgo.Session, cardsMessage *discordgo.Message, authorID string, cards []bson.M) []DropWin {
	var winners []DropWin
	emojis := []string{"1️⃣", "2️⃣", "3️⃣"}
	for i, card := range cards {
		var winner DropWin
		emoji := emojis[i]
		winner.characterID = card["_id"].(string)
		characterInfos, err := mongo.GetCharInfos(winner.characterID)
		if err != nil {
			log.Fatal(err)
		}
		winner.characterName = characterInfos["name"].(string)
		users, err := getUsersWhoReacted(session, cardsMessage.ChannelID, cardsMessage.ID, emoji)
		if err != nil {
			log.Fatal(err)
		}
		if len(users) == 0 {
			continue
		}
		if stringInSlice(authorID, users) {
			winner.winnerID = authorID
		} else {
			rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize the random number generator
			winnerID := users[rand.Intn(len(users))]
			winner.winnerID = winnerID
			winner.opponents = len(users) - 1
		}
		winners = append(winners, winner)
	}
	return winners
}

func NotifyWinners(session *discordgo.Session, channelID string, winners []DropWin) {
	for _, dropWin := range winners {
		charName := dropWin.characterName
		winnerID := dropWin.winnerID
		opponents := dropWin.opponents
		if opponents > 0 {
			session.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> beat %d opponents and won **%s**",
				winnerID, opponents, charName))
		} else {
			session.ChannelMessageSend(channelID, fmt.Sprintf("<@%s> dropped **%s**", winnerID, charName))
		}
	}
}

func AddDropsToInventories(winners []DropWin) {
	for _, dropWin := range winners {
		userID := dropWin.winnerID
		characterID := dropWin.characterID
		mongo.AddToInventory(userID, characterID, 1)
	}
}
