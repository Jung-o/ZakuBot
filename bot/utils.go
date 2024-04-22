package bot

import (
	"ZakuBot/mongo"
	"github.com/nfnt/resize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"math"
	"net/http"
	"os"
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

// CombineDrawnCards downloads the images of the drawn cards, adds a vignette based on the average color of the image,
// writes the character's name and series on the image and combines them into a single image
func CombineDrawnCards(imagesInfo []bson.M) string {
	images := make([]image.Image, len(imagesInfo))
	for i, card := range imagesInfo {
		doc := card["doc"].(primitive.M)
		characterId := doc["characterId"].(string)
		characterInfos, err := mongo.GetCharInfos(characterId)
		if err != nil {
			log.Fatal(err)
		}

		// Download the images
		artworkUrl := doc["artworkUrl"].(string)
		resp, err := http.Get(artworkUrl)
		if err != nil {
			log.Fatal(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		vignettedImage, err := addVignetteBasedOnAverageColor(img)
		if err != nil {
			log.Fatal(err)
		}
		// Convert image to RGBA
		imgRGBA := image.NewRGBA(vignettedImage.Bounds())
		draw.Draw(imgRGBA, imgRGBA.Bounds(), vignettedImage, image.Point{}, draw.Src)
		// Add character's name and series to the image
		name := characterInfos["name"].(string)
		series := characterInfos["series"].(string)
		addLabel(imgRGBA, name, series)
		images[i] = imgRGBA
	}

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

	filePath := "files/combined.jpg"
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	err = jpeg.Encode(file, combined, nil)
	if err != nil {
		log.Fatal(err)
	}

	return filePath
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
