package controldisplay

type Range struct {
	minimum int
	maximum int
}

// Constrain a number to be within a range.
// Returns:
// value: 	if value is between minimum and maximum.
// minimum: if value is less than minimum.
// maximum: if value is greater than maximum.
func (r *Range) Constrain(value int) int {
	if value > r.maximum {
		return r.maximum
	}
	if value < r.minimum {
		return r.minimum
	}
	return value
}

func NewRange(minimum int, maximum int) Range {
	if minimum > maximum {
		panic("invalid range parameters - minimum > maximum")
	}
	return Range{minimum: minimum, maximum: maximum}
}

// MapRange Re-maps a number from one range to another.
func MapRange(value int, valueRange Range, desiredRange Range) int {
	return (value-valueRange.minimum)*(desiredRange.maximum-desiredRange.minimum)/(valueRange.maximum-valueRange.minimum) + desiredRange.minimum
}
