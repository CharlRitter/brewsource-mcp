package brewing

import (
	"errors"
	"math"
)

// ErrInvalidInput is returned when calculation inputs are invalid.
var ErrInvalidInput = errors.New("invalid input for brewing calculation")

// GravityPoint represents a gravity measurement in points (e.g., 1.050 = 50 points).
type GravityPoint float64

// ToSG converts gravity points to specific gravity.
func (gp GravityPoint) ToSG() float64 {
	return 1.0 + float64(gp)/1000.0
}

// FromSG converts specific gravity to gravity points.
func FromSG(sg float64) GravityPoint {
	return GravityPoint((sg - 1.0) * 1000.0)
}

// ABVCalculation represents alcohol calculation methods.
type ABVCalculation struct {
	OriginalGravity float64
	FinalGravity    float64
}

// SimpleABV calculates ABV using the simple formula: (OG - FG) * 131.25.
func (a ABVCalculation) SimpleABV() (float64, error) {
	if a.OriginalGravity <= a.FinalGravity {
		return 0, ErrInvalidInput
	}

	if a.OriginalGravity < 1.0 || a.FinalGravity < 0.990 {
		return 0, ErrInvalidInput
	}

	return (a.OriginalGravity - a.FinalGravity) * 131.25, nil
}

// This is more accurate for higher gravity beers.
func (a ABVCalculation) AlternativeABV() (float64, error) {
	if a.OriginalGravity <= a.FinalGravity {
		return 0, ErrInvalidInput
	}

	if a.OriginalGravity < 1.0 || a.FinalGravity < 0.990 {
		return 0, ErrInvalidInput
	}

	return (a.OriginalGravity - a.FinalGravity) / 7.45 * a.FinalGravity, nil
}

// HopAddition represents a hop addition in brewing.
type HopAddition struct {
	AlphaAcid float64 // Alpha acid percentage (0-20 typical range)
	Amount    float64 // Amount in ounces
	BoilTime  int     // Boil time in minutes
}

// CalculateIBU calculates IBU using the Tinseth formula
// Reference: https://www.realbeer.com/hops/research.html
func (h HopAddition) CalculateIBU(batchSize float64, originalGravity float64) (float64, error) {
	if h.AlphaAcid <= 0 || h.Amount <= 0 || batchSize <= 0 || originalGravity < 1.0 {
		return 0, ErrInvalidInput
	}

	utilizationFactor := h.calculateUtilization(originalGravity)

	// Tinseth IBU formula: (AA% * Weight * Utilization * 74.89) / Volume
	// Note: AlphaAcid is already a percentage (e.g., 5.5 for 5.5%)
	ibu := (h.AlphaAcid * h.Amount * utilizationFactor * 74.89) / batchSize

	return ibu, nil
}

// calculateUtilization calculates hop utilization based on boil time and gravity.
func (h HopAddition) calculateUtilization(originalGravity float64) float64 {
	// Bigness factor accounts for the effect of gravity on utilization
	bignessFactor := 1.65 * math.Pow(0.000125, originalGravity-1.0)

	// Boil time factor
	boilTimeFactor := (1.0 - math.Exp(-0.04*float64(h.BoilTime))) / 4.15

	return bignessFactor * boilTimeFactor
}

// SRMCalculation represents color calculation for beer.
type SRMCalculation struct {
	GrainBill []GrainAddition
	BatchSize float64 // Batch size in gallons
}

// GrainAddition represents a grain addition with color contribution.
type GrainAddition struct {
	Weight   float64 // Weight in pounds
	Lovibond float64 // Lovibond color rating
}

// Using the Morey equation for more accuracy.
func (s SRMCalculation) CalculateSRM() (float64, error) {
	if s.BatchSize <= 0 {
		return 0, ErrInvalidInput
	}

	var mcu float64 // Malt Color Units

	for _, grain := range s.GrainBill {
		if grain.Weight < 0 || grain.Lovibond < 0 {
			return 0, ErrInvalidInput
		}
		mcu += (grain.Weight * grain.Lovibond) / s.BatchSize
	}

	// Morey equation: SRM = 1.4922 * (MCU^0.6859)
	srm := 1.4922 * math.Pow(mcu, 0.6859)

	return srm, nil
}

// SRMToEBC converts SRM to EBC (European Brewery Convention) color scale.
func SRMToEBC(srm float64) float64 {
	return srm * 1.97
}

// EBCToSRM converts EBC to SRM color scale.
func EBCToSRM(ebc float64) float64 {
	return ebc / 1.97
}

// EfficiencyCalculation represents brewhouse efficiency calculations.
type EfficiencyCalculation struct {
	GrainBill       []GrainAddition
	MeasuredGravity float64
	BatchSize       float64
	GrainPotentials map[string]float64 // Potential extract for different grains
}

// CalculateEfficiency calculates brewhouse efficiency.
func (e EfficiencyCalculation) CalculateEfficiency() (float64, error) {
	if e.BatchSize <= 0 || e.MeasuredGravity < 1.0 {
		return 0, ErrInvalidInput
	}

	var potentialPoints float64

	for _, grain := range e.GrainBill {
		if grain.Weight < 0 {
			return 0, ErrInvalidInput
		}

		// Default potential extract if not specified (typical base malt)
		potential := 1.037 // Default potential
		if len(e.GrainPotentials) > 0 {
			// In a real implementation, you'd match grain types to potentials
			potential = 1.037 // Simplified for this example
		}

		grainPoints := float64(FromSG(potential)) * grain.Weight
		potentialPoints += float64(grainPoints)
	}

	potentialGravity := (potentialPoints / e.BatchSize / 1000.0) + 1.0
	measuredPoints := FromSG(e.MeasuredGravity)
	potentialPointsTotal := FromSG(potentialGravity)

	efficiency := float64(measuredPoints) / float64(potentialPointsTotal) * 100.0

	return efficiency, nil
}

// WaterChemistry represents basic water chemistry calculations.
type WaterChemistry struct {
	Calcium     float64 // ppm
	Magnesium   float64 // ppm
	Sodium      float64 // ppm
	Chloride    float64 // ppm
	Sulfate     float64 // ppm
	Bicarbonate float64 // ppm
}

// CalculateRA calculates Residual Alkalinity.
func (w WaterChemistry) CalculateRA() float64 {
	// Simplified RA calculation: RA = HCO3- - (Ca2+/1.4 + Mg2+/1.7)
	ra := w.Bicarbonate - (w.Calcium/1.4 + w.Magnesium/1.7)
	return ra
}

// This ratio affects beer flavor: higher ratios emphasize hop character.
func (w WaterChemistry) CalculateSulfateToChlorideRatio() float64 {
	if w.Chloride == 0 {
		return 0 // Avoid division by zero
	}
	return w.Sulfate / w.Chloride
}

// YeastPitchingRate represents yeast pitching calculations.
type YeastPitchingRate struct {
	BatchSize       float64 // Gallons
	OriginalGravity float64
	YeastViability  float64 // Percentage (0-100)
}

// Returns cells in billions.
func (y YeastPitchingRate) CalculatePitchingRate() (float64, error) {
	if y.BatchSize <= 0 || y.OriginalGravity < 1.0 || y.YeastViability <= 0 || y.YeastViability > 100 {
		return 0, ErrInvalidInput
	}

	// Standard pitching rate for ales: 0.75 million cells/mL/°Plato
	// For lagers: 1.5 million cells/mL/°Plato

	// Convert gravity to Plato
	plato := GravityToPlato(y.OriginalGravity)

	// Volume in mL (1 gallon = 3785.41 mL)
	volumeML := y.BatchSize * 3785.41

	// Calculate required cells for ales (in millions)
	requiredCellsMillions := 0.75 * volumeML * plato

	// Convert to billions
	requiredCellsBillions := requiredCellsMillions / 1000.0

	// Adjust for viability
	adjustedCells := requiredCellsBillions / (y.YeastViability / 100.0)

	return adjustedCells, nil
}

// GravityToPlato converts specific gravity to degrees Plato.
func GravityToPlato(gravity float64) float64 {
	// Balling/Plato approximation
	return (-1.0 * 616.868) + (1111.14 * gravity) - (630.272 * gravity * gravity) + (135.997 * gravity * gravity * gravity)
}

// PlatoToGravity converts degrees Plato to specific gravity.
func PlatoToGravity(plato float64) float64 {
	// Approximation formula
	return 1.0 + (plato / (258.6 - ((plato / 258.2) * 227.1)))
}

// AttenuationCalculation represents attenuation calculations.
type AttenuationCalculation struct {
	OriginalGravity float64
	FinalGravity    float64
}

// ApparentAttenuation calculates apparent attenuation.
func (a AttenuationCalculation) ApparentAttenuation() (float64, error) {
	if a.OriginalGravity <= a.FinalGravity || a.OriginalGravity < 1.0 {
		return 0, ErrInvalidInput
	}

	ogPoints := FromSG(a.OriginalGravity)
	fgPoints := FromSG(a.FinalGravity)

	attenuation := float64(ogPoints-fgPoints) / float64(ogPoints) * 100.0

	return attenuation, nil
}

// RealAttenuation calculates real attenuation (accounting for alcohol's lower density).
func (a AttenuationCalculation) RealAttenuation() (float64, error) {
	if a.OriginalGravity <= a.FinalGravity || a.OriginalGravity < 1.0 {
		return 0, ErrInvalidInput
	}

	// Real extract approximation
	realExtract := (0.1808 * float64(FromSG(a.OriginalGravity))) + (0.8192 * float64(FromSG(a.FinalGravity)))
	realAttenuation := (float64(FromSG(a.OriginalGravity)) - realExtract) / float64(FromSG(a.OriginalGravity)) * 100.0

	return realAttenuation, nil
}
