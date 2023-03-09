package controldisplay

type RangeConstraint struct {
	minimum int
	maximum int
}

// Constrain a number to be within a range.
func (r *RangeConstraint) Constrain(value int) int {
	if value > r.maximum {
		return r.maximum
	}
	if value < r.minimum {
		return r.minimum
	}
	return value
}

func NewRangeConstraint(minimum int, maximum int) RangeConstraint {
	if minimum > maximum {
		panic("invalid range parameters - minimum > maximum")
	}
	return RangeConstraint{minimum: minimum, maximum: maximum}
}

// MapRange Re-maps a number from one range to another.
func MapRange(value int, valueRange RangeConstraint, desiredRange RangeConstraint) int {
	return (value-valueRange.minimum)*(desiredRange.maximum-desiredRange.minimum)/(valueRange.maximum-valueRange.minimum) + desiredRange.minimum
}
