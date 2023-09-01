// Copyright ©2020 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hplot

import (
	"go-hep.org/x/hep/hplot/htex"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// Figure creates a new figure from a plot and options.
// Figure returns a value implementing the Drawer interface.
func Figure(p Drawer, opts ...FigOption) *Fig {
	fig := &Fig{
		Plot:  p,
		Latex: htex.NoopHandler{},
		DPI:   float64(vgimg.DefaultDPI),
	}
	for _, opt := range opts {
		opt(fig)
	}
	return fig
}

// FigOption allows to customize the creation of figures.
type FigOption func(fig *Fig)

// Border specifies the borders' sizes, the space between the
// end of the plot image (PDF, PNG, ...) and the actual plot.
type Border struct {
	Left   vg.Length
	Right  vg.Length
	Bottom vg.Length
	Top    vg.Length
}

// WithBorder allows to specify the borders' sizes, the space between the
// end of the plot image (PDF, PNG, ...) and the actual plot.
func WithBorder(b Border) FigOption {
	return func(fig *Fig) {
		fig.Border = b
	}
}

// WithLatexHandler allows to enable the automatic generation of PDFs from .tex files.
// To enable the automatic generation of PDFs, use DefaultHandler:
//
//	WithLatexHandler(htex.DefaultHandler)
func WithLatexHandler(h htex.Handler) FigOption {
	return func(fig *Fig) {
		fig.Latex = h
	}
}

// WithDPI allows to modify the default DPI of a plot.
func WithDPI(dpi float64) FigOption {
	return func(fig *Fig) {
		fig.DPI = dpi
	}
}

// WithLegend enables the display of a legend on the righthand-side of a plot.
func WithLegend(l Legend) FigOption {
	return func(fig *Fig) {
		leg := l
		fig.Legend = &leg
	}
}

// Fig is a figure, holding a plot and figure-level customizations.
type Fig struct {
	// Plot is a gonum/plot.Plot like value.
	Plot Drawer

	// Legend displays a legend on the righthand-side of the plot.
	Legend *Legend

	// Border specifies the borders' sizes, the space between the
	// end of the plot image (PDF, PNG, ...) and the actual plot.
	Border Border

	// Latex handles the generation of PDFs from .tex files.
	// The default is to use htex.NoopHandler (a no-op).
	// To enable the automatic generation of PDFs, use DefaultHandler:
	//  p := hplot.Wrap(plt)
	//  p.Latex = htex.DefaultHandler
	Latex htex.Handler

	// DPI is the dot-per-inch for PNG,JPEG,... plots.
	DPI float64
}

func (fig *Fig) Draw(dc draw.Canvas) {
	vgtexBorder(dc)

	dc = draw.Crop(dc,
		fig.Border.Left, -fig.Border.Right,
		fig.Border.Bottom, -fig.Border.Top,
	)

	if fig.Legend != nil {
		var (
			r      = fig.Legend.Rectangle(dc)
			width  = r.Max.X - r.Min.X
			height vg.Length
		)
		// adjust the legend down a little.
		switch p := fig.Plot.(type) {
		case *plot.Plot:
			height = p.Title.TextStyle.FontExtents().Height
		case *Plot:
			height = p.Title.TextStyle.FontExtents().Height
		}
		fig.Legend.YOffs = -height

		fig.Legend.Draw(dc)

		// carve up space for the legend.
		dc = draw.Crop(dc, 0, -width-vg.Millimeter, 0, 0)
	}

	fig.Plot.Draw(dc)
}

var (
	_ Drawer = (*Fig)(nil)
)
