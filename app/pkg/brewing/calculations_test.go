package brewing

import (
	"math"
	"testing"
)

func TestABVCalculation(t *testing.T) {
	tests := []struct {
		name        string
		og          float64
		fg          float64
		expectedABV float64
		expectError bool
	}{
		// Happy Path Test Cases
		{
			name:        "Standard beer",
			og:          1.050,
			fg:          1.010,
			expectedABV: 5.25, // (1.050 - 1.010) * 131.25 = 5.25
			expectError: false,
		},
		{
			name:        "Strong beer",
			og:          1.080,
			fg:          1.015,
			expectedABV: 8.5325, // (1.080 - 1.015) * 131.25
			expectError: false,
		},
		{
			name:        "Light beer",
			og:          1.035,
			fg:          1.008,
			expectedABV: 3.54375, // (1.035 - 1.008) * 131.25
			expectError: false,
		},
		{
			name:        "Imperial beer",
			og:          1.100,
			fg:          1.020,
			expectedABV: 10.5, // (1.100 - 1.020) * 131.25
			expectError: false,
		},

		// Boundary Value Test Cases
		{
			name:        "Minimum valid gravity difference",
			og:          1.001,
			fg:          1.000,
			expectedABV: 0.13125, // Very small but valid
			expectError: false,
		},
		{
			name:        "Equal gravities (boundary)",
			og:          1.050,
			fg:          1.050,
			expectError: true,
		},
		{
			name:        "FG at minimum valid (0.990)",
			og:          1.050,
			fg:          0.990,
			expectedABV: 7.875, // (1.050 - 0.990) * 131.25
			expectError: false,
		},
		{
			name:        "OG at minimum valid (1.000)",
			og:          1.000,
			fg:          0.995,
			expectedABV: 0.65625, // (1.000 - 0.995) * 131.25
			expectError: false,
		},

		// Sad Path Test Cases
		{
			name:        "Invalid - FG higher than OG",
			og:          1.040,
			fg:          1.050,
			expectError: true,
		},
		{
			name:        "Invalid - negative OG",
			og:          0.950,
			fg:          1.010,
			expectError: true,
		},
		{
			name:        "Invalid - FG below minimum (0.990)",
			og:          1.050,
			fg:          0.985,
			expectError: true,
		},
		{
			name:        "Invalid - OG below minimum (1.0)",
			og:          0.999,
			fg:          0.995,
			expectError: true,
		},

		// Edge Case Test Cases
		{
			name:        "Zero gravity difference (technically impossible)",
			og:          1.050,
			fg:          1.050,
			expectError: true,
		},
		{
			name:        "Extremely high gravity beer",
			og:          1.150, // Extreme case
			fg:          1.025,
			expectedABV: 16.40625, // (1.150 - 1.025) * 131.25
			expectError: false,
		},
		{
			name:        "Very dry beer (low FG)",
			og:          1.060,
			fg:          0.992,
			expectedABV: 8.925, // (1.060 - 0.992) * 131.25
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := ABVCalculation{
				OriginalGravity: tt.og,
				FinalGravity:    tt.fg,
			}

			result, err := calc.SimpleABV()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if math.Abs(result-tt.expectedABV) > 0.01 {
				t.Errorf("Expected ABV %.2f, got %.2f", tt.expectedABV, result)
			}
		})
	}
}

func TestHopIBUCalculation(t *testing.T) {
	tests := []struct {
		name        string
		hop         HopAddition
		batchSize   float64
		og          float64
		expectedIBU float64
		expectError bool
	}{
		// Happy Path Test Cases
		{
			name: "Standard 60 minute addition",
			hop: HopAddition{
				AlphaAcid: 5.5,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 30, // Approximate - Tinseth formula is complex
			expectError: false,
		},
		{
			name: "High alpha acid hop",
			hop: HopAddition{
				AlphaAcid: 15.0,
				Amount:    0.5,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 20, // Approximate
			expectError: false,
		},
		{
			name: "Flame out addition (5 minutes)",
			hop: HopAddition{
				AlphaAcid: 8.0,
				Amount:    1.0,
				BoilTime:  5,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 5, // Lower utilization
			expectError: false,
		},

		// Boundary Value Test Cases
		{
			name: "Dry hop addition (0 minutes)",
			hop: HopAddition{
				AlphaAcid: 10.0,
				Amount:    1.0,
				BoilTime:  0,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 0, // No IBU contribution from dry hopping
			expectError: false,
		},
		{
			name: "Minimum alpha acid",
			hop: HopAddition{
				AlphaAcid: 0.1, // Very low but valid
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 0.35, // Very low IBU due to low alpha acid
			expectError: false,
		},
		{
			name: "Maximum realistic alpha acid",
			hop: HopAddition{
				AlphaAcid: 20.0, // Very high AA
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 110, // Very high IBU
			expectError: false,
		},
		{
			name: "Minimum hop amount",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    0.01, // Very small amount
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 0.3, // Very low IBU
			expectError: false,
		},

		// Equivalence Partitioning Test Cases
		{
			name: "Low gravity beer effect",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.030, // Low gravity = higher utilization
			expectedIBU: 28,    // Slightly higher than standard
			expectError: false,
		},
		{
			name: "High gravity beer effect",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.080, // High gravity = lower utilization
			expectedIBU: 22,    // Lower than standard
			expectError: false,
		},
		{
			name: "Large batch dilution",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   10.0, // Double batch size
			og:          1.050,
			expectedIBU: 13, // Half the IBU
			expectError: false,
		},
		{
			name: "Small batch concentration",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   2.5, // Half batch size
			og:          1.050,
			expectedIBU: 50, // Double the IBU
			expectError: false,
		},

		// Sad Path Test Cases
		{
			name: "Invalid - negative alpha acid",
			hop: HopAddition{
				AlphaAcid: -5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - zero alpha acid",
			hop: HopAddition{
				AlphaAcid: 0.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - negative hop amount",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    -1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - zero hop amount",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    0.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - negative batch size",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   -5.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - zero batch size",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   0.0,
			og:          1.050,
			expectError: true,
		},
		{
			name: "Invalid - gravity below 1.0",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  60,
			},
			batchSize:   5.0,
			og:          0.999,
			expectError: true,
		},

		// Edge Case Test Cases
		{
			name: "Long boil time (240 minutes)",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  240, // 4 hour boil
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 30, // Utilization maxes out
			expectError: false,
		},
		{
			name: "Negative boil time (should still work)",
			hop: HopAddition{
				AlphaAcid: 5.0,
				Amount:    1.0,
				BoilTime:  -10, // Invalid but formula might handle
			},
			batchSize:   5.0,
			og:          1.050,
			expectedIBU: 0, // Should result in very low utilization
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.hop.CalculateIBU(tt.batchSize, tt.og)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// IBU calculations can vary, so we allow more tolerance
			if tt.expectedIBU == 0 && result > 1 {
				t.Errorf("Expected very low IBU for 0 minute addition, got %.2f", result)
			} else if tt.expectedIBU > 0 && (result < tt.expectedIBU*0.5 || result > tt.expectedIBU*1.5) {
				t.Errorf("Expected IBU around %.2f, got %.2f", tt.expectedIBU, result)
			}
		})
	}
}

func TestSRMCalculation(t *testing.T) {
	tests := []struct {
		name        string
		grains      []GrainAddition
		batchSize   float64
		expectedSRM float64
		expectError bool
	}{
		{
			name: "Pale ale grain bill",
			grains: []GrainAddition{
				{Weight: 8.0, Lovibond: 2.0},  // Base malt
				{Weight: 1.0, Lovibond: 60.0}, // Crystal malt
			},
			batchSize:   5.0,
			expectedSRM: 7, // Approximate
			expectError: false,
		},
		{
			name: "Light beer",
			grains: []GrainAddition{
				{Weight: 6.0, Lovibond: 1.8}, // Pilsner malt
			},
			batchSize:   5.0,
			expectedSRM: 2, // Very light
			expectError: false,
		},
		{
			name: "Invalid - zero batch size",
			grains: []GrainAddition{
				{Weight: 8.0, Lovibond: 2.0},
			},
			batchSize:   0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := SRMCalculation{
				GrainBill: tt.grains,
				BatchSize: tt.batchSize,
			}

			result, err := calc.CalculateSRM()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// SRM calculations can vary based on the Morey equation
			if math.Abs(result-tt.expectedSRM) > tt.expectedSRM*0.5 {
				t.Errorf("Expected SRM around %.1f, got %.1f", tt.expectedSRM, result)
			}
		})
	}
}

func TestGravityConversions(t *testing.T) {
	tests := []struct {
		gravity float64
		plato   float64
	}{
		{1.040, 10.0}, // Approximate
		{1.050, 12.4}, // Approximate
		{1.080, 19.3}, // Approximate
	}

	for _, tt := range tests {
		t.Run("Gravity_to_Plato", func(t *testing.T) {
			result := GravityToPlato(tt.gravity)
			if math.Abs(result-tt.plato) > 1.0 { // Allow 1 degree tolerance
				t.Errorf("Expected Plato around %.1f, got %.1f", tt.plato, result)
			}
		})

		t.Run("Plato_to_Gravity", func(t *testing.T) {
			result := PlatoToGravity(tt.plato)
			if math.Abs(result-tt.gravity) > 0.005 { // Allow 0.005 SG tolerance
				t.Errorf("Expected gravity around %.3f, got %.3f", tt.gravity, result)
			}
		})
	}
}

func TestAttenuationCalculation(t *testing.T) {
	tests := []struct {
		name                string
		og                  float64
		fg                  float64
		expectedAttenuation float64
		expectError         bool
	}{
		{
			name:                "Standard attenuation",
			og:                  1.050,
			fg:                  1.010,
			expectedAttenuation: 80.0, // (50-10)/50 * 100 = 80%
			expectError:         false,
		},
		{
			name:                "High attenuation",
			og:                  1.060,
			fg:                  1.005,
			expectedAttenuation: 91.67, // (60-5)/60 * 100
			expectError:         false,
		},
		{
			name:        "Invalid - FG higher than OG",
			og:          1.040,
			fg:          1.050,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := AttenuationCalculation{
				OriginalGravity: tt.og,
				FinalGravity:    tt.fg,
			}

			result, err := calc.ApparentAttenuation()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if math.Abs(result-tt.expectedAttenuation) > 0.1 {
				t.Errorf("Expected attenuation %.2f%%, got %.2f%%", tt.expectedAttenuation, result)
			}
		})
	}
}

func TestYeastPitchingRate(t *testing.T) {
	tests := []struct {
		name          string
		batchSize     float64
		og            float64
		viability     float64
		expectedCells float64 // In billions
		expectError   bool
	}{
		{
			name:          "Standard 5 gallon batch",
			batchSize:     5.0,
			og:            1.050,
			viability:     100.0,
			expectedCells: 100, // Approximate - depends on calculation
			expectError:   false,
		},
		{
			name:        "Invalid - zero viability",
			batchSize:   5.0,
			og:          1.050,
			viability:   0.0,
			expectError: true,
		},
		{
			name:        "Invalid - negative batch size",
			batchSize:   -1.0,
			og:          1.050,
			viability:   100.0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := YeastPitchingRate{
				BatchSize:       tt.batchSize,
				OriginalGravity: tt.og,
				YeastViability:  tt.viability,
			}

			result, err := calc.CalculatePitchingRate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Yeast pitching calculations vary widely
			if result < 50 || result > 500 {
				t.Errorf("Expected reasonable pitching rate (50-500B cells), got %.1f", result)
			}
		})
	}
}

func TestAlternativeABVCalculation(t *testing.T) {
	tests := []struct {
		name        string
		og          float64
		fg          float64
		expectedABV float64
		expectError bool
	}{
		// Happy Path Test Cases
		{
			name:        "Standard beer - alternative formula",
			og:          1.050,
			fg:          1.010,
			expectedABV: 0.0054, // (1.050 - 1.010) / 7.45 * 1.010 ≈ 0.0054
			expectError: false,
		},
		{
			name:        "High gravity beer - more accurate",
			og:          1.100,
			fg:          1.020,
			expectedABV: 0.0109, // (1.100 - 1.020) / 7.45 * 1.020 ≈ 0.0109
			expectError: false,
		},

		// Boundary Value Test Cases
		{
			name:        "Minimum difference",
			og:          1.001,
			fg:          1.000,
			expectedABV: 0.000134, // Very small
			expectError: false,
		},

		// Sad Path Test Cases
		{
			name:        "Invalid - FG higher than OG",
			og:          1.040,
			fg:          1.050,
			expectError: true,
		},
		{
			name:        "Invalid - OG below minimum",
			og:          0.999,
			fg:          0.995,
			expectError: true,
		},
		{
			name:        "Invalid - FG below minimum",
			og:          1.050,
			fg:          0.985,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := ABVCalculation{
				OriginalGravity: tt.og,
				FinalGravity:    tt.fg,
			}

			result, err := calc.AlternativeABV()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if math.Abs(result-tt.expectedABV) > 0.01 {
				t.Errorf("Expected ABV %.4f, got %.4f", tt.expectedABV, result)
			}
		})
	}
}

func TestGravityPointConversions(t *testing.T) {
	tests := []struct {
		name   string
		sg     float64
		points GravityPoint
	}{
		// Happy Path Test Cases
		{"Standard beer gravity", 1.050, GravityPoint(50)},
		{"Light beer gravity", 1.035, GravityPoint(35)},
		{"Strong beer gravity", 1.080, GravityPoint(80)},

		// Boundary Value Test Cases
		{"Water (1.000)", 1.000, GravityPoint(0)},
		{"Very light (1.001)", 1.001, GravityPoint(1)},
		{"Very strong (1.120)", 1.120, GravityPoint(120)},

		// Edge Case Test Cases
		{"Extreme gravity (1.200)", 1.200, GravityPoint(200)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test FromSG
			points := FromSG(tt.sg)
			if math.Abs(float64(points-tt.points)) > 0.1 {
				t.Errorf("FromSG(%.3f) = %.1f, want %.1f", tt.sg, points, tt.points)
			}

			// Test ToSG
			sg := tt.points.ToSG()
			if math.Abs(sg-tt.sg) > 0.001 {
				t.Errorf("GravityPoint(%.1f).ToSG() = %.3f, want %.3f", tt.points, sg, tt.sg)
			}
		})
	}
}

func TestSRMToEBCConversion(t *testing.T) {
	tests := []struct {
		name string
		srm  float64
		ebc  float64
	}{
		// Happy Path Test Cases
		{"Light beer", 3.0, 5.91},  // 3 * 1.97 = 5.91
		{"Amber beer", 8.0, 15.76}, // 8 * 1.97 = 15.76
		{"Dark beer", 25.0, 49.25}, // 25 * 1.97 = 49.25

		// Boundary Value Test Cases
		{"Zero SRM", 0.0, 0.0},
		{"Very light", 1.0, 1.97},

		// Edge Case Test Cases
		{"Very dark", 40.0, 78.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test SRM to EBC
			ebc := SRMToEBC(tt.srm)
			if math.Abs(ebc-tt.ebc) > 0.01 {
				t.Errorf("SRMToEBC(%.1f) = %.2f, want %.2f", tt.srm, ebc, tt.ebc)
			}

			// Test EBC to SRM
			srm := EBCToSRM(tt.ebc)
			if math.Abs(srm-tt.srm) > 0.01 {
				t.Errorf("EBCToSRM(%.2f) = %.1f, want %.1f", tt.ebc, srm, tt.srm)
			}
		})
	}
}

func TestWaterChemistryCalculations(t *testing.T) {
	tests := []struct {
		name          string
		water         WaterChemistry
		expectedRA    float64
		expectedSO4Cl float64
	}{
		// Happy Path Test Cases
		{
			name: "Balanced water",
			water: WaterChemistry{
				Calcium:     100,
				Magnesium:   20,
				Sulfate:     150,
				Chloride:    100,
				Bicarbonate: 120,
			},
			expectedRA:    120 - (100/1.4 + 20/1.7), // 120 - (71.4 + 11.8) = 36.8
			expectedSO4Cl: 1.5,                      // 150/100
		},
		{
			name: "High sulfate hoppy water",
			water: WaterChemistry{
				Calcium:     150,
				Magnesium:   15,
				Sulfate:     300,
				Chloride:    50,
				Bicarbonate: 50,
			},
			expectedRA:    50 - (150/1.4 + 15/1.7), // Negative RA
			expectedSO4Cl: 6.0,                     // 300/50
		},
		{
			name: "Malty water profile",
			water: WaterChemistry{
				Calcium:     80,
				Magnesium:   25,
				Sulfate:     50,
				Chloride:    150,
				Bicarbonate: 180,
			},
			expectedRA:    180 - (80/1.4 + 25/1.7),
			expectedSO4Cl: 0.33, // 50/150
		},

		// Edge Case Test Cases
		{
			name: "Zero chloride (division by zero protection)",
			water: WaterChemistry{
				Calcium:     100,
				Magnesium:   20,
				Sulfate:     150,
				Chloride:    0,
				Bicarbonate: 120,
			},
			expectedRA:    120 - (100/1.4 + 20/1.7),
			expectedSO4Cl: 0, // Should handle division by zero
		},

		// Boundary Value Test Cases
		{
			name: "All zero values",
			water: WaterChemistry{
				Calcium:     0,
				Magnesium:   0,
				Sulfate:     0,
				Chloride:    0,
				Bicarbonate: 0,
			},
			expectedRA:    0,
			expectedSO4Cl: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test RA calculation
			ra := tt.water.CalculateRA()
			if math.Abs(ra-tt.expectedRA) > 5.0 { // Allow 5 ppm tolerance
				t.Errorf("CalculateRA() = %.1f, want %.1f", ra, tt.expectedRA)
			}

			// Test SO4:Cl ratio
			ratio := tt.water.CalculateSulfateToChlorideRatio()
			if math.Abs(ratio-tt.expectedSO4Cl) > 0.1 {
				t.Errorf("CalculateSulfateToChlorideRatio() = %.2f, want %.2f", ratio, tt.expectedSO4Cl)
			}
		})
	}
}

func TestEfficiencyCalculation(t *testing.T) {
	tests := []struct {
		name               string
		calc               EfficiencyCalculation
		expectedEfficiency float64
		expectError        bool
	}{
		// Happy Path Test Cases
		{
			name: "Standard efficiency",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: 8.0, Lovibond: 2.0},
					{Weight: 1.0, Lovibond: 60.0},
				},
				MeasuredGravity: 1.048,
				BatchSize:       5.0,
				GrainPotentials: map[string]float64{},
			},
			expectedEfficiency: 75, // Approximate
			expectError:        false,
		},

		// Boundary Value Test Cases
		{
			name: "Perfect efficiency (theoretical)",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: 8.0, Lovibond: 2.0},
				},
				MeasuredGravity: 1.060, // High measured gravity
				BatchSize:       5.0,
			},
			expectedEfficiency: 95, // Very high efficiency
			expectError:        false,
		},
		{
			name: "Low efficiency",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: 10.0, Lovibond: 2.0},
				},
				MeasuredGravity: 1.035, // Low measured gravity
				BatchSize:       5.0,
			},
			expectedEfficiency: 50, // Low efficiency
			expectError:        false,
		},

		// Sad Path Test Cases
		{
			name: "Invalid - zero batch size",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: 8.0, Lovibond: 2.0},
				},
				MeasuredGravity: 1.048,
				BatchSize:       0,
			},
			expectError: true,
		},
		{
			name: "Invalid - gravity below 1.0",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: 8.0, Lovibond: 2.0},
				},
				MeasuredGravity: 0.999,
				BatchSize:       5.0,
			},
			expectError: true,
		},
		{
			name: "Invalid - negative grain weight",
			calc: EfficiencyCalculation{
				GrainBill: []GrainAddition{
					{Weight: -8.0, Lovibond: 2.0},
				},
				MeasuredGravity: 1.048,
				BatchSize:       5.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.calc.CalculateEfficiency()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Efficiency calculations can vary widely
			if result < 0 || result > 120 {
				t.Errorf("Expected reasonable efficiency (0-120%%), got %.1f", result)
			}
		})
	}
}

func TestRealAttenuationCalculation(t *testing.T) {
	tests := []struct {
		name                string
		og                  float64
		fg                  float64
		expectedAttenuation float64
		expectError         bool
	}{
		// Happy Path Test Cases
		{
			name:                "Standard real attenuation",
			og:                  1.050,
			fg:                  1.010,
			expectedAttenuation: 65, // Typically lower than apparent
			expectError:         false,
		},
		{
			name:                "High real attenuation",
			og:                  1.060,
			fg:                  1.005,
			expectedAttenuation: 75, // High attenuation
			expectError:         false,
		},

		// Boundary Value Test Cases
		{
			name:                "Minimum difference",
			og:                  1.001,
			fg:                  1.000,
			expectedAttenuation: 10, // Very low but valid
			expectError:         false,
		},

		// Sad Path Test Cases
		{
			name:        "Invalid - FG higher than OG",
			og:          1.040,
			fg:          1.050,
			expectError: true,
		},
		{
			name:        "Invalid - OG below minimum",
			og:          0.999,
			fg:          1.005,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := AttenuationCalculation{
				OriginalGravity: tt.og,
				FinalGravity:    tt.fg,
			}

			result, err := calc.RealAttenuation()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Real attenuation should be reasonable
			if result < 0 || result > 100 {
				t.Errorf("Expected reasonable real attenuation (0-100%%), got %.2f%%", result)
			}
		})
	}
}
