package glyph

// TextBlockC is a multi-line text element with word wrapping.
// Unlike TextView, it has no Layer and no scrolling — it reports its
// natural wrapped height to the layout system and renders inline.
// Use inside ScrollView or VBox for flowing multi-line content.
type TextBlockC struct {
	content  any // string or *string
	style    Style
	styleDyn any
	fgDyn    any
	bgDyn    any
}

// TextBlock creates a multi-line text display with word wrapping.
// Content is re-wrapped at the width assigned by the layout.
func TextBlock(content any) TextBlockC {
	return TextBlockC{content: content}
}

func (t TextBlockC) Style(s any) TextBlockC {
	switch v := s.(type) {
	case Style:
		t.style = v
	default:
		t.styleDyn = v
	}
	return t
}

func (t TextBlockC) FG(c any) TextBlockC {
	switch v := c.(type) {
	case Color:
		t.style.FG = v
	default:
		t.fgDyn = v
	}
	return t
}

func (t TextBlockC) BG(c any) TextBlockC {
	switch v := c.(type) {
	case Color:
		t.style.BG = v
	default:
		t.bgDyn = v
	}
	return t
}

func (t TextBlockC) Bold() TextBlockC  { t.style.Attr |= AttrBold; return t }
func (t TextBlockC) Dim() TextBlockC   { t.style.Attr |= AttrDim; return t }
func (t TextBlockC) Italic() TextBlockC { t.style.Attr |= AttrItalic; return t }
