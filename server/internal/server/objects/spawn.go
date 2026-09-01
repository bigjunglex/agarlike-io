package objects

import "math/rand/v2"

var getPlayerPos = func(p *Player) (float64, float64) { return p.X, p.Y }
var getPlayerRad = func(p *Player) float64 { return p.Radius }
var getSporePos = func(s *Spore) (float64, float64) { return s.X, s.Y }
var getSporeRad = func(s *Spore) float64 { return s.Radius }

func GetSpawnCoords(
	radius float64,
	playersToAvoid *SharedCollection[*Player],
	sporesToAvoid *SharedCollection[*Spore],
) (float64, float64) {
	bound := 2000.0
	const maxTries int = 25

	tries := 0

	for {
		x := bound * (2 * rand.Float64())
		y := bound * (2 * rand.Float64())

		if !isTooClose(x, y, radius, playersToAvoid, getPlayerPos, getPlayerRad) &&
			!isTooClose(x, y, radius, sporesToAvoid, getSporePos, getSporeRad) {
			return x, y
		}

		tries++
		if tries > maxTries {
			bound *= 2
			tries = 0
		}
	}

}

func isTooClose[T any](
	x float64,
	y float64,
	r float64,
	objects *SharedCollection[T],
	getPos func(T) (float64, float64),
	getRad func(T) float64,
) bool {
	if objects == nil {
		return false
	}

	check := false
	objects.ForEach(func(_ uint64, o T) {
		if check {
			return
		}
		oX, oY := getPos(o)
		oR := getRad(o)
		dX := oX - x
		dY := oY - y
		dSq := dX*dX + dY*dY

		if dSq <= (r+oR)*(r+oR) {
			check = true
			return
		}
	})

	return check
}
