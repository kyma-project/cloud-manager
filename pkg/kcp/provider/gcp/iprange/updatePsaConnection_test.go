package iprange

import (
	"testing"

	"google.golang.org/api/servicenetworking/v1"
)

func TestDoesConnectionMatchPeeringRanges(t *testing.T) {
	tests := []struct {
		name             string
		connectionRanges []string
		peeringIpRanges  []string
		expectedMatch    bool
		description      string
	}{
		{
			name:             "exact match - single range",
			connectionRanges: []string{"cm-abc-123"},
			peeringIpRanges:  []string{"cm-abc-123"},
			expectedMatch:    true,
			description:      "Connection has exactly the same single range",
		},
		{
			name:             "exact match - multiple ranges",
			connectionRanges: []string{"cm-abc-123", "cm-def-456", "cm-ghi-789"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456", "cm-ghi-789"},
			expectedMatch:    true,
			description:      "Connection has exactly the same multiple ranges",
		},
		{
			name:             "exact match - different order",
			connectionRanges: []string{"cm-def-456", "cm-abc-123", "cm-ghi-789"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456", "cm-ghi-789"},
			expectedMatch:    true,
			description:      "Connection has same ranges but in different order",
		},
		{
			name:             "no match - missing range in connection",
			connectionRanges: []string{"cm-abc-123", "cm-def-456"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456", "cm-ghi-789"},
			expectedMatch:    false,
			description:      "Connection is missing a range that should be added",
		},
		{
			name:             "no match - extra range in connection",
			connectionRanges: []string{"cm-abc-123", "cm-def-456", "cm-ghi-789"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456"},
			expectedMatch:    false,
			description:      "Connection has an extra range that should be removed",
		},
		{
			name:             "no match - completely different ranges",
			connectionRanges: []string{"cm-xxx-111", "cm-yyy-222"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456"},
			expectedMatch:    false,
			description:      "Connection has completely different ranges",
		},
		{
			name:             "no match - empty connection",
			connectionRanges: []string{},
			peeringIpRanges:  []string{"cm-abc-123"},
			expectedMatch:    false,
			description:      "Connection is empty but should have ranges",
		},
		{
			name:             "match - both empty",
			connectionRanges: []string{},
			peeringIpRanges:  []string{},
			expectedMatch:    true,
			description:      "Both connection and desired ranges are empty",
		},
		{
			name:             "no match - nil connection",
			connectionRanges: nil,
			peeringIpRanges:  []string{"cm-abc-123"},
			expectedMatch:    false,
			description:      "Connection ranges are nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &State{
				serviceConnection: &servicenetworking.Connection{
					ReservedPeeringRanges: tt.connectionRanges,
				},
				peeringIpRanges: tt.peeringIpRanges,
			}

			// Test with nil connection
			if tt.connectionRanges == nil {
				state.serviceConnection = nil
			}

			result := state.DoesConnectionMatchPeeringRanges()

			if result != tt.expectedMatch {
				t.Errorf("%s: expected match=%v, got match=%v\n  Description: %s\n  Connection ranges: %v\n  Desired ranges: %v",
					tt.name, tt.expectedMatch, result, tt.description, tt.connectionRanges, tt.peeringIpRanges)
			}
		})
	}
}

func TestUpdatePsaConnectionIdempotency(t *testing.T) {
	// This test documents the expected behavior:
	// updatePsaConnection should only call PatchServiceConnection when ranges differ

	tests := []struct {
		name             string
		connectionRanges []string
		peeringIpRanges  []string
		shouldUpdate     bool
		description      string
	}{
		{
			name:             "should NOT update - ranges match",
			connectionRanges: []string{"cm-abc-123", "cm-def-456"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456"},
			shouldUpdate:     false,
			description:      "Connection already has correct ranges, no update needed",
		},
		{
			name:             "should NOT update - ranges match different order",
			connectionRanges: []string{"cm-def-456", "cm-abc-123"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456"},
			shouldUpdate:     false,
			description:      "Same ranges but different order, no update needed",
		},
		{
			name:             "SHOULD update - need to add range",
			connectionRanges: []string{"cm-abc-123"},
			peeringIpRanges:  []string{"cm-abc-123", "cm-def-456"},
			shouldUpdate:     true,
			description:      "Need to add cm-def-456 to connection",
		},
		{
			name:             "SHOULD update - need to remove range",
			connectionRanges: []string{"cm-abc-123", "cm-def-456"},
			peeringIpRanges:  []string{"cm-abc-123"},
			shouldUpdate:     true,
			description:      "Need to remove cm-def-456 from connection",
		},
		{
			name:             "SHOULD update - need to delete connection",
			connectionRanges: []string{"cm-abc-123"},
			peeringIpRanges:  []string{},
			shouldUpdate:     true,
			description:      "No ranges left, should delete connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &State{
				serviceConnection: &servicenetworking.Connection{
					ReservedPeeringRanges: tt.connectionRanges,
				},
				peeringIpRanges: tt.peeringIpRanges,
			}

			shouldUpdate := !state.DoesConnectionMatchPeeringRanges()

			if shouldUpdate != tt.shouldUpdate {
				t.Errorf("%s: expected shouldUpdate=%v, got shouldUpdate=%v\n  Description: %s\n  Connection ranges: %v\n  Desired ranges: %v",
					tt.name, tt.shouldUpdate, shouldUpdate, tt.description, tt.connectionRanges, tt.peeringIpRanges)
			}
		})
	}
}
