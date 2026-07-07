package tui

// PayloadGeometry owns every payload dimension.
// Nothing outside PayloadGeometry performs layout calculations.
type payloadGeometry struct {
	AvailableWidth int
	KeyWidth       int
	ValueWidth     int
	BodyWidth      int
	BodyHeight     int
}

func calculatePayloadGeometry(availableWidth int) payloadGeometry {
	if availableWidth < 20 {
		availableWidth = 20
	}

	innerWidth := availableWidth - 2
	if innerWidth < 10 {
		innerWidth = 10
	}

	headerWidth := max(10, innerWidth-5)
	keyW := max(10, headerWidth*4/10)
	valueW := max(10, headerWidth-keyW)

	bodyWidth := max(10, availableWidth-len(indentNested)-2)
	bodyHeight := 3

	return payloadGeometry{
		AvailableWidth: availableWidth,
		KeyWidth:       keyW,
		ValueWidth:     valueW,
		BodyWidth:      bodyWidth,
		BodyHeight:     bodyHeight,
	}
}
