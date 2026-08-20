package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type minSizeObject struct {
	widget.BaseWidget
	content fyne.CanvasObject
	size    fyne.Size
}

// minSize wraps a canvas object to enforce a minimum width and height without altering its inherent layout properties.
func minSize(content fyne.CanvasObject, size fyne.Size) fyne.CanvasObject {
	object := &minSizeObject{content: content, size: size}
	object.ExtendBaseWidget(object)
	return object
}

// CreateRenderer creates and returns a standard Fyne widget renderer for the custom minSize layout container.
func (object *minSizeObject) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(object.content)
}

// MinSize calculates the minimum necessary dimensions by combining the content's minimum size and the forced size.
func (object *minSizeObject) MinSize() fyne.Size {
	return object.content.MinSize().Max(object.size)
}

type responsiveGridLayout struct {
	minCellWidth float32
	maxColumns   int
}

// responsiveGrid creates a dynamic grid layout that automatically adjusts column counts based on available screen width.
func responsiveGrid(minCellWidth float32, maxColumns int, objects ...fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&responsiveGridLayout{minCellWidth: minCellWidth, maxColumns: maxColumns}, objects...)
}

// Layout arranges the child objects in a grid, dynamically wrapping to new rows based on the calculated column count.
func (grid *responsiveGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	visible := visibleObjects(objects)
	columns := grid.columnsFor(size.Width, len(visible))
	if columns < 1 {
		return
	}

	padding := theme.Padding()
	rows := (len(visible) + columns - 1) / columns
	cellWidth := (size.Width - float32(columns-1)*padding) / float32(columns)
	if cellWidth < 0 {
		cellWidth = size.Width
	}

	maxMinHeight := float32(0)
	for _, object := range visible {
		if h := object.MinSize().Height; h > maxMinHeight {
			maxMinHeight = h
		}
	}

	cellHeight := (size.Height - float32(rows-1)*padding) / float32(rows)
	if cellHeight < maxMinHeight {
		cellHeight = maxMinHeight
	}

	for index, object := range visible {
		row := index / columns
		column := index % columns
		object.Move(fyne.NewPos(float32(column)*(cellWidth+padding), float32(row)*(cellHeight+padding)))
		object.Resize(fyne.NewSize(cellWidth, cellHeight))
	}
}

// MinSize computes the absolute minimum dimensions required to display the grid with the minimal allowed column count.
func (grid *responsiveGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	visible := visibleObjects(objects)
	if len(visible) == 0 {
		return fyne.NewSize(0, 0)
	}

	cell := fyne.NewSize(grid.minCellWidth, 0)
	for _, object := range visible {
		cell = cell.Max(object.MinSize())
	}

	columns := min(grid.maxColumns, len(visible))
	if columns < 1 {
		columns = 1
	}
	rows := (len(visible) + columns - 1) / columns
	padding := theme.Padding()
	return fyne.NewSize(
		cell.Width,
		float32(rows)*cell.Height+float32(rows-1)*padding,
	)
}

// columnsFor calculates the optimal number of columns that can fit into the given width without shrinking below the minimum.
func (grid *responsiveGridLayout) columnsFor(width float32, itemCount int) int {
	if itemCount == 0 {
		return 0
	}

	padding := theme.Padding()
	columns := int((width + padding) / (grid.minCellWidth + padding))
	columns = max(1, columns)
	columns = min(columns, grid.maxColumns)
	return min(columns, itemCount)
}

type maxWidthLayout struct {
	width float32
}

// maxWidth centers a canvas object and restricts it from growing beyond a specified maximum horizontal width.
func maxWidth(content fyne.CanvasObject, width float32) fyne.CanvasObject {
	return container.New(&maxWidthLayout{width: width}, content)
}

// Layout centers the child objects horizontally if the container exceeds the constrained maximum width.
func (layout *maxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, object := range objects {
		if !object.Visible() {
			continue
		}
		width := min(size.Width, layout.width)
		x := (size.Width - width) / 2
		object.Move(fyne.NewPos(x, 0))
		object.Resize(fyne.NewSize(width, size.Height))
	}
}

// MinSize aggregates and returns the largest minimum size requirements from all visible child objects.
func (layout *maxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, object := range objects {
		if object.Visible() {
			minSize = minSize.Max(object.MinSize())
		}
	}
	return minSize
}

// visibleObjects filters a given slice of Fyne canvas objects, returning only those currently marked as visible.
func visibleObjects(objects []fyne.CanvasObject) []fyne.CanvasObject {
	visible := make([]fyne.CanvasObject, 0, len(objects))
	for _, object := range objects {
		if object.Visible() {
			visible = append(visible, object)
		}
	}
	return visible
}
