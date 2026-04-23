package geo

import "testing"

func TestHaversineDistance(t *testing.T) {
	tests := []struct {
		name         string
		lat1, lon1   float64
		lat2, lon2   float64
		wantMeters   int64
		tolerancePct float64 // acceptable relative error
	}{
		{
			name: "same point",
			lat1: 55.75, lon1: 37.62,
			lat2: 55.75, lon2: 37.62,
			wantMeters:   0,
			tolerancePct: 0,
		},
		{
			// Moscow Kremlin → Saint Petersburg roughly 635 km
			name: "Moscow to Saint Petersburg",
			lat1: 55.7520, lon1: 37.6175,
			lat2: 59.9343, lon2: 30.3351,
			wantMeters:   634_000,
			tolerancePct: 2.0,
		},
		{
			// Equator, 1 degree longitude ≈ 111 320 m
			name: "equator 1 degree",
			lat1: 0, lon1: 0,
			lat2: 0, lon2: 1,
			wantMeters:   111_320,
			tolerancePct: 0.2,
		},
		{
			// Antipodal points ≈ half Earth circumference ≈ 20 015 km
			name: "antipodal",
			lat1: 0, lon1: 0,
			lat2: 0, lon2: 180,
			wantMeters:   20_015_000,
			tolerancePct: 0.1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HaversineDistance(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
			if tc.wantMeters == 0 {
				if got != 0 {
					t.Fatalf("want 0, got %d", got)
				}
				return
			}
			diff := got - tc.wantMeters
			if diff < 0 {
				diff = -diff
			}
			pct := float64(diff) / float64(tc.wantMeters) * 100
			if pct > tc.tolerancePct {
				t.Fatalf("want ~%d m, got %d m (%.2f%% error, tolerance %.1f%%)",
					tc.wantMeters, got, pct, tc.tolerancePct)
			}
		})
	}
}
