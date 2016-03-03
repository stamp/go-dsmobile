package gui

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/disintegration/gift"
	"github.com/stamp/go-dsmobile/types"
)

type Package struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type FileItem struct {
	Src      string   `json:"src"`
	Thumb    string   `json:"thumb"`
	Height   int      `json:"h"`
	Width    int      `json:"w"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

func (c *connection) sendFileList() {
	//list := make([]FileItem, 0)
	files, _ := ioutil.ReadDir("storage")

	for _, file := range files {
		fileInfo, _ := os.Stat("storage/" + file.Name())
		if fileInfo.IsDir() {
			continue
		}

		width, height := getImageDimension("storage/" + file.Name())

		thumb, err := getThumb("storage/", file.Name())
		if err != nil {
			continue
		}
		//		list = append(list,
		item := FileItem{
			Src:    "storage/" + file.Name(),
			Thumb:  thumb,
			Height: height,
			Width:  width,
		}

		pkg := Package{Type: "file", Data: item}
		data, _ := json.Marshal(pkg)

		// TODO: panic: send on closed channel
		select {
		case c.send <- data:
		default:
		}
	}
}

func (c *connection) sendCategories(categories *types.Categories) {
	pkg := Package{Type: "categories", Data: categories.GetAll()}
	data, _ := json.Marshal(pkg)

	// TODO: panic: send on closed channel
	select {
	case c.send <- data:
	default:
	}
}

func getThumb(path, imagePath string) (string, error) { // {{{
	dir, _ := filepath.Split(imagePath)

	thumbFilename := path + "thumbs/" + imagePath

	if _, err := os.Stat(path + "thumbs/" + dir); err != nil {
		err := os.MkdirAll(path+"thumbs/"+dir, 077)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(thumbFilename); err == nil {
		return thumbFilename, nil
	}

	err := makeThumb(path+imagePath, thumbFilename)

	if err != nil {
		return "", err
	}

	return thumbFilename, nil
}                                                  // }}}
func makeThumb(source, destination string) error { // {{{
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	g := gift.New(
		gift.ResizeToFill(200, 200, gift.LanczosResampling, gift.CenterAnchor),
	)

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	// write new image to file
	dst := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)
	return jpeg.Encode(out, dst, &jpeg.Options{100})
}                                                     // }}}
func getImageDimension(imagePath string) (int, int) { // {{{
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
} // }}}
