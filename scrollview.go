package glyph

// ScrollViewC wraps children in a scrollable layer. Children are laid out
// using normal glyph components, rendered into an offscreen buffer once,
// and then blitted each frame. Re-renders only when the viewport width changes.
//
// Usage:
//
//	ScrollView.Grow(1)(
//	    Text("Hello").Bold(),
//	    SpaceH(1),
//	    Text("World"),
//	)
type ScrollViewC struct {
	layer    *Layer
	children []any
	flexGrow float32
	margin   [4]int16

	// cached sub-template, rebuilt on width change
	childTmpl *Template
}

type ScrollViewFn func(children ...any) *ScrollViewC

// ScrollView creates a scrollable container for its children.
var ScrollView ScrollViewFn = func(children ...any) *ScrollViewC {
	sv := &ScrollViewC{
		layer:    NewLayer(),
		children: children,
	}
	sv.layer.Render = sv.render
	return sv
}

func (f ScrollViewFn) Grow(g any) ScrollViewFn {
	return func(children ...any) *ScrollViewC {
		sv := f(children...)
		switch val := g.(type) {
		case float32:
			sv.flexGrow = val
		case float64:
			sv.flexGrow = float32(val)
		case int:
			sv.flexGrow = float32(val)
		}
		return sv
	}
}

func (f ScrollViewFn) Margin(all int16) ScrollViewFn {
	return func(children ...any) *ScrollViewC {
		sv := f(children...)
		sv.margin = [4]int16{all, all, all, all}
		return sv
	}
}

func (f ScrollViewFn) MarginVH(v, h int16) ScrollViewFn {
	return func(children ...any) *ScrollViewC {
		sv := f(children...)
		sv.margin = [4]int16{v, h, v, h}
		return sv
	}
}

// Ref captures a reference to this ScrollView via a callback,
// allowing inline assignment during view composition.
func (sv *ScrollViewC) Ref(f func(*ScrollViewC)) *ScrollViewC { f(sv); return sv }

// Layer returns the underlying layer for scroll control wiring.
func (sv *ScrollViewC) Layer() *Layer {
	return sv.layer
}

// Refresh forces re-render of children on the next frame.
// Call when the content has changed.
func (sv *ScrollViewC) Refresh() {
	sv.layer.lastRenderWidth = 0
}

func (t *Template) compileScrollViewC(v *ScrollViewC, parent int16, depth int) int16 {
	layerView := LayerView(v.layer).Grow(v.flexGrow)
	if v.margin != [4]int16{} {
		layerView = layerView.MarginTRBL(v.margin[0], v.margin[1], v.margin[2], v.margin[3])
	}
	return t.compileLayerViewC(layerView, parent, depth)
}

func (sv *ScrollViewC) render() {
	w := sv.layer.ViewportWidth()
	if w <= 0 {
		return
	}

	// build child template once (or rebuild if children change via Refresh)
	if sv.childTmpl == nil {
		sv.childTmpl = Build(VBox(sv.children...))
	}

	// run layout at the viewport width to get natural content height
	sv.childTmpl.distributeWidths(int16(w), nil)
	sv.childTmpl.layout(32000) // large height so children aren't constrained

	// read the natural content height from the root op
	contentH := int(sv.childTmpl.geom[0].ContentH)
	if contentH < sv.layer.ViewportHeight() {
		contentH = sv.layer.ViewportHeight()
	}

	// render into a buffer sized to the actual content
	buf := NewBuffer(w, contentH)
	buf.defaultStyle = sv.layer.defaultStyle
	buf.Clear()
	sv.childTmpl.render(buf, 0, 0, int16(w))

	sv.layer.SetBuffer(buf)
}
