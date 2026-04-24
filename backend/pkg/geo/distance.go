package geo

import "math"

const earthRadiusMeters = 6_371_000

// HaversineDistance returns the straight-line distance in metres between two WGS84 coordinates.
func HaversineDistance(lat1, lon1, lat2, lon2 float64) int64 {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int64(math.Round(earthRadiusMeters * c))
}
