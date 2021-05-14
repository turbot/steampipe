package execute

type DimensionColorMap map[string]map[string]uint8

func newDimensionColorMap(e *ExecutionTree) DimensionColorMap {
	colorMap := make(DimensionColorMap)
	colorMap.populate(e)
	return colorMap
}

func (c *DimensionColorMap) populate(e *ExecutionTree) {
	var prevColor uint8
	for _, run := range e.controlRuns {
		for _, r := range run.Result.Rows {
			for _, d := range r.Dimensions {
				if !c.hasDimensionValue(d) {
					prevColor = c.addDimensionValue(d, prevColor)
				}
			}
		}
	}
}

func (c DimensionColorMap) hasDimensionValue(dimension Dimension) bool {
	dimensionMap, ok := c[dimension.Key]
	var gotValue bool
	if ok {
		// so we have a dimension map for this dimension
		// - do we have a dimension color for this property value?
		_, gotValue = dimensionMap[dimension.Value]
	}
	return gotValue
}

func (c DimensionColorMap) addDimensionValue(d Dimension, prevColor uint8) uint8 {
	// do we have a dimension map for this dimension property?

	if c[d.Key] == nil {
		c[d.Key] = make(map[string]uint8)
	}

	color := c.getNextColor(prevColor)
	// store the color keyed by property VALUE
	c[d.Key][d.Value] = color
	return color
}

func (c *DimensionColorMap) getNextColor(prevColor uint8) uint8 {
	// color codes range from 16-231:  6 × 6 × 6 cube (216 colors): 16 + 36 × r + 6 × g + b (0 ≤ r, g, b ≤ 5)
	const interval = 50
	if prevColor == 0 {
		return 32
	}

	color := prevColor + interval
	if color > 231 {
		color = 32
	}
	return color
}

//ALARM: is pretty insecure .......................................................................................................................................................................................................................................................................................................   partition 10000 us-east-2 3335354343537
//1.1 Maintain current contact details ...................................................................................................................................................................................................................................................................................................................   1 /  1 [=         ]
