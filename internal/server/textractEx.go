package server

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/fogleman/gg"
)

const imgPath = "./cmd/api/angled_cards.jpg"

func (s *Server) extractText(ctx context.Context) ([]types.Block, error) {
	// read the image file into a byte slice
	imageBytes, err := os.ReadFile(imgPath)
	if err != nil {
		log.Println("Error reading image file:", err)
		return nil, err
	}

	// prepare the input
	input := &textract.DetectDocumentTextInput{
		Document: &types.Document{
			Bytes: imageBytes,
		},
	}

	result, err := s.client.DetectDocumentText(ctx, input)
	if err != nil {
		log.Println("Error detecting text:", err)
		return nil, err
	}

	// Collect polygons for low-confidence WORD blocks.
	var polygons [][]types.Point
	for _, block := range result.Blocks {
		if block.BlockType == "LINE" {
			// fmt.Printf("text: %s\tconf: %d", *block.Text, *block.Confidence)
			spew.Dump(block)
		}

		if block.BlockType == "WORD" && block.Confidence != nil && *block.Confidence < 80.0 {
			if block.Geometry.Polygon != nil && len(block.Geometry.Polygon) > 0 {
				polygons = append(polygons, block.Geometry.Polygon)
			}
		}
	}

	if len(polygons) > 0 {
		err = annotateImage(imgPath, polygons)
		if err != nil {
			log.Println("Error saving annotated image")
		}
	}

	return result.Blocks, err
}

func annotateImage(imagePath string, polygons [][]types.Point) error {
	// load the original image
	img, err := gg.LoadImage(imagePath)
	if err != nil {
		return fmt.Errorf("loading image: %w", err)
	}
	bounds := img.Bounds()
	imgWidth := float64(bounds.Dx())
	imgHeight := float64(bounds.Dy())

	// create a drawing context
	dc := gg.NewContextForImage(img)
	dc.SetLineWidth(2)
	dc.SetRGB255(255, 0, 0) // red

	const padding = 5.0 // 5 pixel padding to expand the polygon

	// draw each polygon
	for _, poly := range polygons {
		if len(poly) == 0 {
			continue
		}

		// create a new polygon with padding
		paddedPoly := padPolygon(poly, padding, float32(imgWidth), float32(imgHeight))

		// draw the polygon
		dc.MoveTo(float64(paddedPoly[0].X), float64(paddedPoly[0].Y))
		for _, pt := range paddedPoly[1:] {
			dc.LineTo(float64(pt.X), float64(pt.Y))
		}

		dc.ClosePath()
		dc.Stroke()
	}

	// create the new filename and save
	ext := filepath.Ext(imagePath)
	base := imagePath[:len(imagePath)-len(ext)]
	outputPath := fmt.Sprintf("%s_for_review%s", base, ext)
	return dc.SavePNG(outputPath)
}

// padPolygon takes a polygon (slice of Textract Points in normalized coordinates)
// and returns a new polygon with each vertex moved outward by 'padding' pixels.
func padPolygon(poly []types.Point, padding, imgWidth, imgHeight float32) []types.Point {
	var points []types.Point
	var sumX, sumY float32
	// Convert normalized coordinates to pixel coordinates and find the middle
	for _, p := range poly {
		x := p.X * imgWidth
		y := p.Y * imgHeight
		points = append(points, types.Point{X: x, Y: y})
		sumX += x
		sumY += y
	}
	n := float32(len(points))
	middleX := sumX / n
	middleY := sumY / n

	// go through each point and push it outward
	var padded []types.Point
	for _, pt := range points {
		// figure out direction from middle
		dx := pt.X - middleX
		dy := pt.Y - middleY

		// how far is this point from middle?
		dist := math.Hypot(float64(dx), float64(dy))

		// if it's right on the middle, just keep it
		if dist == 0 {
			padded = append(padded, pt)
		} else {
			// push the point outward by padding amount
			// this makes the polygon bigger
			factor := (float32(dist) + padding) / float32(dist)

			// calculate new position
			newX := middleX + dx*factor
			newY := middleY + dy*factor

			// add the new point
			padded = append(padded, types.Point{X: newX, Y: newY})
		}
	}
	return padded
}

// BoundingBox represents the normalized bounding box values.
// type BoundingBox struct {
// 	Height float64
// 	Left   float64
// 	Top    float64
// 	Width  float64
// }

// Bounding boxes
// func annotateImage(imagePath string, boxes []types.BoundingBox) error {
// 	// Load the original image.
// 	img, err := gg.LoadImage(imagePath)
// 	if err != nil {
// 		return fmt.Errorf("loading image: %w", err)
// 	}
// 	bounds := img.Bounds()
// 	imgWidth := float32(bounds.Dx())
// 	imgHeight := float32(bounds.Dy())

// 	// Create a drawing context.
// 	dc := gg.NewContextForImage(img)
// 	dc.SetLineWidth(2)
// 	// Set red color.
// 	dc.SetRGB255(255, 0, 0)

// 	// Draw each bounding box.
// 	for _, box := range boxes {
// 		// Convert normalized coordinates to pixel values.
// 		x := box.Left * imgWidth
// 		y := box.Top * imgHeight
// 		w := box.Width * imgWidth
// 		h := box.Height * imgHeight

// 		dc.DrawRectangle(float64(x), float64(y), float64(w), float64(h))
// 		dc.Stroke()
// 	}

// 	// Create the new filename.
// 	ext := filepath.Ext(imagePath)
// 	base := imagePath[:len(imagePath)-len(ext)]
// 	outputPath := fmt.Sprintf("%s_for_review%s", base, ext)

// 	// Save the annotated image.
// 	return dc.SavePNG(outputPath)
// }
