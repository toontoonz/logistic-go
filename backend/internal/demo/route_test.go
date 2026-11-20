package demo

import "testing"

func TestPointAlongStaysOnSegment(t *testing.T) {
	path := []latLng{{0, 0}, {0, 1}}
	mid := pointAlong(path, haversine(path[0], path[1])/2)
	if mid.lat < -0.01 || mid.lat > 0.01 {
		t.Fatalf("lat = %v", mid.lat)
	}
	if mid.lng < 0.4 || mid.lng > 0.6 {
		t.Fatalf("lng = %v", mid.lng)
	}
	start := pointAlong(bangkokRoute, 0)
	if start != bangkokRoute[0] {
		t.Fatalf("start = %+v", start)
	}
	end := pointAlong(bangkokRoute, pathLength(bangkokRoute)+10)
	if end != bangkokRoute[len(bangkokRoute)-1] {
		t.Fatalf("end = %+v", end)
	}
}
