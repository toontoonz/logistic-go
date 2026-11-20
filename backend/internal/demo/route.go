// Package demo moves a driver along a fixed Bangkok route so the map
// can be shown without a phone. The points match frontend/lib/routes.ts.
package demo

import "math"

type latLng struct {
	lat float64
	lng float64
}

// bangkokRoute is Siam → Thong Lo, the same polyline the web simulator draws.
var bangkokRoute = []latLng{
	{13.7462, 100.5347},
	{13.7398, 100.5412},
	{13.7308, 100.5416},
	{13.7278, 100.5518},
	{13.7372, 100.5602},
	{13.7306, 100.5694},
	{13.7241, 100.5788},
}

func haversine(a, b latLng) float64 {
	const earth = 6371000.0
	p1 := a.lat * math.Pi / 180
	p2 := b.lat * math.Pi / 180
	dp := (b.lat - a.lat) * math.Pi / 180
	dl := (b.lng - a.lng) * math.Pi / 180
	h := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return 2 * earth * math.Asin(math.Min(1, math.Sqrt(h)))
}

func pathLength(path []latLng) float64 {
	total := 0.0
	for i := 1; i < len(path); i++ {
		total += haversine(path[i-1], path[i])
	}
	return total
}

func pointAlong(path []latLng, distance float64) latLng {
	if len(path) == 0 {
		return latLng{lat: 13.736, lng: 100.555}
	}
	if len(path) == 1 || distance <= 0 {
		return path[0]
	}
	left := distance
	for i := 1; i < len(path); i++ {
		segment := haversine(path[i-1], path[i])
		if left <= segment {
			t := 0.0
			if segment != 0 {
				t = left / segment
			}
			return latLng{
				lat: path[i-1].lat + (path[i].lat-path[i-1].lat)*t,
				lng: path[i-1].lng + (path[i].lng-path[i-1].lng)*t,
			}
		}
		left -= segment
	}
	return path[len(path)-1]
}
