package interpolation

import (
	"context"
	"earth/nasa"
	"earth/shared"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"strings"
	"sync"
)

const width = float64(2048)
const halfWidth = width / 2
const height = float64(2048)
const halfHeight = height / 2

func loadNasaImage(im nasa.ImageMeta) (image.Image, error) {
	path := nasa.BuildImageFilePath(im)
	return shared.LoadImage(path)
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func computeDistanceFromEarth(im nasa.ImageMeta) float64 {
	x := im.SatellitePosition.X
	y := im.SatellitePosition.Y
	z := im.SatellitePosition.Z
	return math.Sqrt((x * x) + (y * y) + (z * z))
}

// To determine the scale of the earth, we can create a relationship
// between the radius of the earth in one of the photos with the distance the satellite was
// from the earth when the photo was taken.

// The pixels count below was measured manually for the selected sample image.
const sampleEarthRadiusInPixels = 791

// Taken from "2020-09-04 00:03:41" metadata
var sampleSatellitePosition = struct {
	x float64
	y float64
	z float64
}{
	x: -1326191.794585,
	y: 698719.827083,
	z: 229500.537516,
}

func computeEarthScale(distanceFromEarth float64) float64 {
	sx := sampleSatellitePosition.x
	sy := sampleSatellitePosition.y
	sz := sampleSatellitePosition.z
	sampleImageDistanceFromEarth := math.Sqrt((sx * sx) + (sy * sy) + (sz * sz))
	scaleFactor := float64((sampleEarthRadiusInPixels / halfWidth) * sampleImageDistanceFromEarth)
	return scaleFactor / distanceFromEarth
}

func computeEarthScaleFromImage(im nasa.ImageMeta) float64 {
	return computeEarthScale(computeDistanceFromEarth(im))
}

func lngDiff(ln1, ln2 float64) float64 {
	return math.Mod((ln1-ln2)+360, 360)
}

// Given a point on the projected virtual sphere,
// get the x,y coordinates of the same point on the sphere within
// one of the real images.
// http://www.movable-type.co.uk/scripts/latlong.html
func latLngToCoordinates(projectedLat, projectedLng, realImageLat, realImageLng float64) (float64, float64) {
	x := math.Sin(projectedLng-realImageLng) * math.Cos(projectedLat)
	y := math.Sin(projectedLat)*math.Cos(realImageLat) - math.Cos(projectedLng-realImageLng)*math.Cos(projectedLat)*math.Sin(realImageLat)
	return x, y
}

// GenerateFrame -
// Given two images, interpolate at the provided lng
// which is expected to fall somewhere between the first image lng
// the second image lng.
func generateFrame(im1, im2 nasa.ImageMeta, lng float64, outFileName string) error {
	centroid1 := im1.CentroidCoordinates
	image1, err := loadNasaImage(im1)
	if err != nil {
		log.Println("error loading image 1", im1.Image)
		return err
	}
	earthScale1 := computeEarthScaleFromImage(im1)

	centroid2 := im2.CentroidCoordinates
	image2, err := loadNasaImage(im2)
	if err != nil {
		log.Println("error loading image 2", im2.Image)
		return err
	}
	earthScale2 := computeEarthScaleFromImage(im2)

	// Each image will carry a weight between 0 and 1 where w1 + w2 = 1
	// These weights can be used to find the lat/lng of the virtual centroid
	// as well as computing the composite pixel value.
	weight1 := lngDiff(centroid1.Lng, lng) / lngDiff(centroid1.Lng, centroid2.Lng)
	weight2 := lngDiff(lng, centroid2.Lng) / lngDiff(centroid1.Lng, centroid2.Lng)

	// The centroid latitude for the virtual sphere in radians
	centroidLatRadians := -degreesToRadians((centroid1.Lat * weight1) + (centroid2.Lat * weight2))
	// The interpolated distance from earth the satellite should be at this point between the two images
	distanceFromEarth := (computeDistanceFromEarth(im1) * weight1) + (computeDistanceFromEarth(im2) * weight2)
	// How large the earth should appear as a percentage of the image width (ex. 0.78)
	earthScale := computeEarthScale(distanceFromEarth)

	// The in-memory synthesized image to be written to disk
	pixelsOut := image.NewRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{int(width), int(height)}})

	for x2D := 0; x2D < int(width); x2D++ {
		for y2D := 0; y2D < int(height); y2D++ {
			// initialze the current pixel as black (deep space!)
			pixelsOut.Set(x2D, y2D, color.Black)
			// scale x approprately for this sphere size
			// we want to fit this vitural sphere to the "unit sphere"
			// where the radius of the surface is 1.
			x2DScaled := (float64(x2D) - halfWidth) / (halfWidth * earthScale)
			y2DScaled := -(float64(y2D) - halfHeight) / (halfHeight * earthScale)

			radius := x2DScaled*x2DScaled + y2DScaled*y2DScaled
			// If we're outside the bounds of the surface of the virtual sphere, we're
			// looking into space, there is no need to interpolate this pixel as it
			// should remain black.
			if radius > 1 {
				continue
			}

			// get 3D coordinates from 2 dimentional x,y
			x3D := float64(x2DScaled)
			y3D := math.Cos(centroidLatRadians)*float64(y2DScaled) - math.Sin(centroidLatRadians)*math.Sqrt(1-radius)
			z3D := math.Sin(centroidLatRadians)*float64(y2DScaled) + math.Cos(centroidLatRadians)*math.Sqrt(1-radius)

			// get lat lng from 3D coordinates using the provided lng
			projectedLat := math.Asin(y3D)
			projectedLng := math.Atan2(x3D, z3D) + degreesToRadians(lng)

			if projectedLng < 0 {
				projectedLng += math.Pi * 2
			}

			// now for each sphere, figure out what 2D x/y the lat/lng corresponds to
			x1, y1 := latLngToCoordinates(projectedLat, projectedLng, degreesToRadians(centroid1.Lat), degreesToRadians(centroid1.Lng))
			x1Scaled := int((x1 * halfWidth * earthScale1) + halfWidth)
			y1Scaled := int(-(y1 * halfHeight * earthScale1) + halfHeight)
			pixel1 := image1.At(x1Scaled, y1Scaled)

			x2, y2 := latLngToCoordinates(projectedLat, projectedLng, degreesToRadians(centroid2.Lat), degreesToRadians(centroid2.Lng))
			x2Scaled := int((x2 * halfWidth * earthScale2) + halfWidth)
			y2Scaled := int(-(y2 * halfHeight * earthScale2) + halfHeight)
			pixel2 := image2.At(x2Scaled, y2Scaled)

			r1, g1, b1, a1 := pixel1.RGBA()
			r2, g2, b2, a2 := pixel2.RGBA()

			pixel := color.RGBA64{
				// weighted RGB
				R: uint16(math.Sqrt(float64(r1*r1)*weight2 + float64(r2*r2)*weight1)),
				G: uint16(math.Sqrt(float64(g1*g1)*weight2 + float64(g2*g2)*weight1)),
				B: uint16(math.Sqrt(float64(b1*b1)*weight2 + float64(b2*b2)*weight1)),
				A: uint16(math.Sqrt(float64(a1*a1)*weight2 + float64(a2*a2)*weight1)),
			}
			pixelsOut.SetRGBA64(x2D, y2D, pixel)
		}
	}

	f, _ := os.Create(outFileName)
	err = png.Encode(f, pixelsOut)
	return err
}

func buildFrameFilePath(im nasa.ImageMeta, frame int) string {
	dateFilePath := nasa.BuildImageFilePath(im)
	return fmt.Sprintf(strings.Split(dateFilePath, "_000.png")[0]+"_%03d.png", frame)
}

// InterpolateImages -
func InterpolateImages(ctx context.Context, ims <-chan nasa.ImageMeta) <-chan bool {
	out := make(chan bool)

	go func() {
		defer func() {
			out <- true
			close(out)
		}()
		const step = 0.5
		var prevIm nasa.ImageMeta

		for {
			select {
			case <-ctx.Done():
				return
			case im, ok := <-ims:
				if !ok {
					return
				}
				if prevIm.Date == "" {
					prevIm = im
					continue
				}

				paths := make([]string, 0)
				wg := sync.WaitGroup{}

				fmt.Println("----------------------")
				fmt.Printf("prev: %s, im: %s\n", prevIm.Date, im.Date)

				originalDiff := lngDiff(prevIm.CentroidCoordinates.Lng, im.CentroidCoordinates.Lng)
				diff := originalDiff
				lng := prevIm.CentroidCoordinates.Lng
				frame := 0
				for true {
					frame++
					// fmt.Println("FRAME: ", frame)
					lng -= step
					if lng < -180 {
						lng += 360
					}
					diff = lngDiff(lng, im.CentroidCoordinates.Lng)
					if diff > originalDiff {
						break
					}
					path := buildFrameFilePath(prevIm, frame)
					paths = append(paths, path)
					if !shared.FileExists(path) {
						wg.Add(1)
						go func(lng float64, path string) {
							err := generateFrame(prevIm, im, lng, path)
							if err != nil {
								log.Fatal(err)
							}
							wg.Done()
						}(lng, path)
					}
				}

				wg.Wait()

				prevIm = im
			}
		}
	}()

	return out
}
