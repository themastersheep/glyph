package glyph_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/kungfusheep/glyph"
)

type termData struct {
	W     int      `json:"w"`
	H     int      `json:"h"`
	Lines []string `json:"lines"`
}

func renderAndPrint(name string, tree any, w, h int) {
	buf := NewBuffer(w, h)
	Build(tree).Execute(buf, int16(w), int16(h))
	fmt.Println(strings.TrimSpace(buf.StringTrimmed()))

	if dir := os.Getenv("GLYPH_PREVIEWS_DIR"); dir != "" {
		td := termData{W: w, H: h}
		for y := 0; y < h; y++ {
			td.Lines = append(td.Lines, buf.GetLineStyled(y))
		}
		data, _ := json.Marshal(td)
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, name+".json"), data, 0644)
	}
}

// Vertical stack.
// Call VBox directly with children to stack them top to bottom.
func ExampleVBoxFn() {
	// example:
	_ = VBox(
		Text("First"),
		Text("Second"),
		Text("Third"),
	)
	// :example

	renderAndPrint("VBoxFn", VBox.Border(BorderRounded).Width(20)(
		Text("Header").Bold().FG(Cyan),
		Text("Body text"),
		Text("Footer").Dim(),
	), 22, 5)
	// Output:
	// ╭──────────────────╮
	// │Header            │
	// │Body text         │
	// │Footer            │
	// ╰──────────────────╯
}

// Template syntax.
// Chain methods to configure, then call as a function with children. Configure once, render many.
func ExampleVBoxFn_chained() {
	// example:
	tree := VBox.Gap(1).Border(BorderRounded)(
		Text("First"),
		Text("Second"),
	)
	// :example

	renderAndPrint("VBoxFn_chained", tree, 20, 6)
	// Output:
	// ╭──────────────────╮
	// │First             │
	// │                  │
	// │Second            │
	// │                  │
	// ╰──────────────────╯
}

// Flex growth.
// Grow fills available space. CascadeStyle passes a style to all descendants.
func ExampleVBoxFn_grow() {
	// example:
	_ = VBox.Grow(1).CascadeStyle(&Style{FG: White})(
		Text("header"),
		Text("content"),
		Text("footer"),
	)
	// :example

	renderAndPrint("VBoxFn_grow", VBox.Border(BorderRounded).Width(30).Height(7)(
		Text("── header ──").Bold().FG(Cyan),
		VBox.Grow(1)(Text("content fills remaining space")),
		Text("── footer ──").Dim(),
	), 32, 9)
	// Output:
	// ╭────────────────────────────╮
	// │── header ──                │
	// │content fills remaining spac│
	// │                            │
	// │                            │
	// │── footer ──                │
	// ╰────────────────────────────╯
}

// Horizontal layout.
// Arrange children side by side with a gap between them.
func ExampleHBoxFn() {
	// example:
	_ = HBox.Gap(2)(
		Text("left"),
		Text("right"),
	)
	// :example

	renderAndPrint("HBoxFn", HBox.Gap(1)(
		VBox.Border(BorderRounded).Grow(1)(Text("Panel A")),
		VBox.Border(BorderRounded).Grow(1)(Text("Panel B")),
		VBox.Border(BorderRounded).Grow(1)(Text("Panel C")),
	), 42, 3)
	// Output:
	// ╭───────────╮ ╭───────────╮ ╭────────────╮
	// │Panel A    │ │Panel B    │ │Panel C     │
	// ╰───────────╯ ╰───────────╯ ╰────────────╯
}

// Sidebar pattern.
// Combine WidthPct and Grow for a fixed sidebar with a flexible main area.
func ExampleHBoxFn_widths() {
	// example:
	_ = HBox(
		VBox.WidthPct(0.3)(Text("sidebar")),
		VBox.Grow(1)(Text("main content")),
	)
	// :example

	renderAndPrint("HBoxFn_widths", HBox(
		VBox.WidthPct(0.3).Border(BorderRounded).Title("nav")(
			Text("Home").Bold(),
			Text("About").Dim(),
			Text("Blog").Dim(),
		),
		VBox.Grow(1).Border(BorderRounded).Title("content")(
			Text("Welcome to the site").Bold().FG(Cyan),
			Text(""),
			Text("Main content area grows to fill.").Dim(),
		),
	), 48, 5)
	// Output:
	// ╭─ nav ──────╮╭─ content ──────────────────────╮
	// │Home        ││Welcome to the site             │
	// │About       ││                                │
	// │Blog        ││Main content area grows to fill.│
	// ╰────────────╯╰────────────────────────────────╯
}

// Basic layering.
// Layer children on top of each other. The last child renders on top.
func ExampleOverlayFn() {
	// example:
	tree := Overlay(
		Text("base content"),
		Text("floating dialog"),
	)
	// :example

	renderAndPrint("OverlayFn", tree, 20, 2)
	// Output:
	// base content
	// floating dialog
}

// Modal dialog.
// Centered dialog with a dimmed backdrop. Size constrains the dialog; Backdrop dims everything behind it.
func ExampleOverlayFn_modal() {
	Overlay.Centered().Backdrop().BackdropFG(BrightBlack).Size(50, 15).BG(Black)(
		VBox.Grow(1)(
			Text("main content"),
		),
		VBox.Border(BorderRounded)(
			Text("confirm delete?"),
			HBox.Gap(2)(Text("[y]es"), Text("[n]o")),
		),
	)
}

// Static text.
// Text that never changes. Pass a string literal directly.
func ExampleText() {
	// example:
	_ = Text("hello world")
	// :example

	renderAndPrint("Text", VBox(
		Text("plain text"),
		Text("bold").Bold(),
		Text("dim").Dim(),
		Text("italic").Italic(),
		Text("coloured").FG(Cyan),
	), 20, 5)
	// Output:
	// plain text
	// bold
	// dim
	// italic
	// coloured
}

// Pointer binding.
// Pass a pointer so the rendered text reflects the current value. Mutate the string, then trigger a render to see it.
func ExampleText_pointer() {
	// example:
	msg := "dynamic"
	_ = Text(&msg)
	// :example

	count := 42
	renderAndPrint("Text_pointer", VBox.Border(BorderRounded).Width(24)(
		HBox.Gap(1)(Text("count:").Dim(), Text(&count).Bold().FG(Cyan)),
		HBox.Gap(1)(Text("label:").Dim(), Text(&msg)),
	), 26, 4)
	// Output:
	// ╭──────────────────────╮
	// │count: 42             │
	// │label: dynamic        │
	// ╰──────────────────────╯
}

// Inline styling.
// Chain style methods. Bold, Dim, Italic, Underline, and FG/BG are all available.
func ExampleText_styled() {
	// example:
	msg := "styled"
	_ = Text(&msg).Bold().FG(Cyan)
	// :example

	renderAndPrint("Text_styled", VBox(
		Text("Bold + Cyan").Bold().FG(Cyan),
		Text("Dim + Italic").Dim().Italic(),
		Text("Underline + Yellow").Underline().FG(Yellow),
		Text("Red on black").FG(Red).BG(Black),
	), 24, 4)
	// Output:
	// Bold + Cyan
	// Dim + Italic
	// Underline + Yellow
	// Red on black
}

// Text input.
// A managed text input with placeholder text shown when empty.
func ExampleInput() {
	Input().Placeholder("Type here...")
}

// Password input.
// Mask hides input characters for sensitive fields.
func ExampleInput_mask() {
	Input().Placeholder("Password").Mask('*')
}

// Navigable list.
// A list over a pointer to a slice. The Render callback receives a pointer to each item for in-place updates.
func ExampleList() {
	items := []string{"Alpha", "Beta", "Gamma"}

	List(&items).Render(func(item *string) any {
		return Text(item)
	}).BindNav("j", "k")
}

// Toggleable list.
// Like List, but each item has a toggleable checkbox. The Done field is toggled in place through the pointer.
func ExampleCheckList() {
	type Task struct {
		Name string
		Done bool
	}
	tasks := []Task{{Name: "Ship it"}, {Name: "Test it"}}

	CheckList(&tasks).Render(func(t *Task) any {
		return Text(&t.Name)
	}).BindNav("j", "k").BindToggle(" ")
}

// Tab bar.
// Bound to an int pointer. The selected index is written when the user picks a tab.
func ExampleTabs() {
	// example:
	var active int
	tabs := []string{"General", "Advanced", "About"}
	tree := Tabs(tabs, &active)
	// :example

	renderAndPrint("Tabs", tree, 40, 1)
	// Output: General  Advanced  About
}

// Custom tab styles.
// Customize appearance with Kind for the shape and per-state styles for active/inactive tabs.
func ExampleTabs_styled() {
	// example:
	var active int
	tabs := []string{"Code", "Preview", "Settings"}
	tree := Tabs(tabs, &active).
		Kind(TabsStyleBox).
		ActiveStyle(Style{FG: Cyan, Attr: AttrBold}).
		InactiveStyle(Style{FG: BrightBlack})
	// :example

	renderAndPrint("Tabs_styled", tree, 40, 3)
	// Output:
	// ┌──────┐  ┌─────────┐  ┌──────────┐
	// │ Code │  │ Preview │  │ Settings │
	// └──────┘  └─────────┘  └──────────┘
}

// Searchable list.
// The extract function tells the filter which string to match against for each item.
func ExampleFilterList() {
	items := []string{"Alpha", "Beta", "Gamma"}

	FilterList(&items, func(s *string) string { return *s }).
		Render(func(item *string) any {
			return Text(item)
		})
}

// Struct table.
// Generate a table from a struct slice. Column options control formatting.
func ExampleAutoTable() {
	// example:
	type Row struct {
		Name string
		CPU  float64
	}
	rows := []Row{{Name: "api", CPU: 42.5}}
	tree := AutoTable(&rows).Column("CPU", Number(1))
	// :example

	renderAndPrint("AutoTable", tree, 40, 5)
	// Output:
	// Name                                CPU
	// api                                42.5
}

// Column formatters.
// Mix formatters: percentages, byte sizes, booleans, and colour-coded changes.
func ExampleAutoTable_columns() {
	// example:
	type Service struct {
		Name   string
		CPU    float64
		Memory uint64
		Active bool
		Growth float64
	}
	rows := []Service{
		{Name: "api", CPU: 82.3, Memory: 1073741824, Active: true, Growth: 0.12},
	}
	tree := AutoTable(&rows).
		Column("CPU", Percent(1)).
		Column("Memory", Bytes()).
		Column("Active", Bool("Yes", "No")).
		Column("Growth", PercentChange(1))
	// :example

	renderAndPrint("AutoTable_columns", tree, 60, 5)
	// Output:
	// Name            CPU       Memory    Active          Growth
	// api           82.3%       1.0 GB     Yes             +0.1%
}

// Currency formatting.
// Format a float as a monetary value with the given symbol and decimal places.
func ExampleCurrency() {
	// example:
	type Invoice struct {
		Item  string
		Price float64
	}
	rows := []Invoice{{Item: "Widget", Price: 42.50}}
	tree := AutoTable(&rows).Column("Price", Currency("$", 2))
	// :example

	renderAndPrint("Currency", tree, 40, 5)
	// Output:
	// Item                              Price
	// Widget                           $42.50
}

// Progress bar.
// Bound to a float64 pointer (0.0–1.0). Width sets the bar length in cells.
func ExampleProgress() {
	var pct int

	Progress(&pct).Width(30).FG(Green)
}

// Animated spinner.
// The frame pointer must be incremented by the caller (e.g. via a ticker goroutine).
func ExampleSpinner() {
	var frame int

	Spinner(&frame).Frames(SpinnerBraille)
}

// Radio group.
// Bound to an int pointer. The selected index updates when the user navigates and presses enter.
func ExampleRadio() {
	// example:
	var selected int
	tree := Radio(&selected, "Small", "Medium", "Large").BindNav("j", "k")
	// :example

	renderAndPrint("Radio", tree, 20, 3)
	// Output:
	// ◉ Small
	// ○ Medium
	// ○ Large
}

// Single checkbox.
// Bound to a bool pointer.
func ExampleCheckbox() {
	// example:
	var agreed bool
	tree := Checkbox(&agreed, "I agree to the terms")
	// :example

	renderAndPrint("Checkbox", tree, 30, 1)
	// Output: ☐ I agree to the terms
}

// Leader line.
// Label on the left, value on the right, dots filling the gap. Both sides read from pointers each frame.
func ExampleLeader() {
	// example:
	label := "Total"
	value := "$42.00"
	tree := Leader(&label, &value)
	// :example

	renderAndPrint("Leader", tree, 20, 1)
	// Output: Total.........$42.00
}

// Mini chart.
// A line chart from a float64 slice pointer. Append values and trigger a render to update.
func ExampleSparkline() {
	// example:
	values := []float64{1, 3, 5, 2, 8, 4}
	tree := Sparkline(&values).Width(20).FG(Green)
	// :example

	renderAndPrint("Sparkline", tree, 25, 1)
	// Output: ▁▃▅▂█▄
}

// Flexible spacer.
// Pushes siblings apart. In a VBox, it pushes content to the top and bottom.
func ExampleSpace() {
	// example:
	tree := VBox(
		Text("header"),
		Space(),
		Text("footer"),
	)
	// :example

	renderAndPrint("Space", tree, 20, 3)
	// Output:
	// header
	//
	// footer
}

// Fixed-height gap.
// Use in VBox to add exact spacing between elements.
func ExampleSpaceH() {
	VBox(
		Text("above"),
		SpaceH(2),
		Text("below"),
	)
}

// Fixed-width gap.
// Use in HBox to add exact spacing between elements.
func ExampleSpaceW() {
	// example:
	tree := HBox(
		Text("left"),
		SpaceW(4),
		Text("right"),
	)
	// :example

	renderAndPrint("SpaceW", tree, 20, 1)
	// Output: left    right
}

// Horizontal divider.
// A line that fills the available width.
func ExampleHRule() {
	// example:
	tree := VBox(
		Text("above"),
		HRule(),
		Text("below"),
	)
	// :example

	renderAndPrint("HRule", tree, 10, 3)
	// Output:
	// above
	// ──────────
	// below
}

// Vertical divider.
// A line that fills the available height.
func ExampleVRule() {
	// example:
	tree := HBox(
		Text("left"),
		VRule(),
		Text("right"),
	)
	// :example

	renderAndPrint("VRule", tree, 15, 1)
	// Output: left│right
}

// Scrollbar.
// Parameters are total content height, visible viewport height, and a pointer to scroll position.
func ExampleScroll() {
	var pos int
	Scroll(100, 20, &pos)
}

// Conditional show.
// Show or hide content based on a bool pointer. The value is checked every frame.
func ExampleIf() {
	// example:
	show := true
	_ = If(&show).Then(Text("visible"))
	// :example

	show2 := true
	renderAndPrint("If", VBox(
		HBox.Gap(1)(Text("true  →").Dim(), If(&show2).Then(Text("✓ visible").FG(Green))),
		HBox.Gap(1)(Text("false →").Dim(), Text("  (nothing rendered)").Dim()),
	), 30, 2)
	// Output:
	// true  → ✓ visible
	// false →   (nothing rendered)
}

// Toggle views.
// Switch between two views based on state.
func ExampleIf_else() {
	// example:
	loggedIn := false
	_ = If(&loggedIn).Then(Text("dashboard")).Else(Text("login"))
	// :example

	t, f := true, false
	renderAndPrint("If_else", VBox(
		HBox.Gap(1)(Text("loggedIn=true  →").Dim(), If(&t).Then(Text("dashboard").FG(Green)).Else(Text("login").FG(Yellow))),
		HBox.Gap(1)(Text("loggedIn=false →").Dim(), If(&f).Then(Text("dashboard").FG(Green)).Else(Text("login").FG(Yellow))),
	), 36, 2)
	// Output:
	// loggedIn=true  → dashboard
	// loggedIn=false → login
}

// Value matching.
// Match against a specific value. Works with strings, ints, or any comparable type.
func ExampleCondition_eq() {
	// example:
	status := "active"
	_ = If(&status).Eq("active").Then(
		Text("online").FG(Green),
	).Else(
		Text("offline").FG(Red),
	)
	// :example

	s1, s2 := "active", "inactive"
	renderAndPrint("Condition_eq", VBox(
		HBox.Gap(1)(Text("status=\"active\"   →").Dim(), If(&s1).Eq("active").Then(Text("● online").FG(Green)).Else(Text("● offline").FG(Red))),
		HBox.Gap(1)(Text("status=\"inactive\" →").Dim(), If(&s2).Eq("active").Then(Text("● online").FG(Green)).Else(Text("● offline").FG(Red))),
	), 40, 2)
	// Output:
	// status="active"   → ● online
	// status="inactive" → ● offline
}

// Multi-way branch.
// Branch on a value with named cases. Default catches unmatched values.
func ExampleSwitch() {
	// example:
	mode := "edit"
	_ = Switch(&mode).
		Case("edit", Text("editing")).
		Case("preview", Text("previewing")).
		Default(Text("idle"))
	// :example

	m1, m2, m3 := "edit", "preview", "other"
	row := func(label string, m *string) any {
		return HBox.Gap(1)(Text(label).Dim(), Switch(m).
			Case("edit", Text("✎ editing").FG(Green)).
			Case("preview", Text("👁 previewing").FG(Cyan)).
			Default(Text("… idle").FG(BrightBlack)))
	}
	renderAndPrint("Switch", VBox(
		row("mode=\"edit\"    →", &m1),
		row("mode=\"preview\" →", &m2),
		row("mode=\"other\"   →", &m3),
	), 36, 3)
	// Output:
	// mode="edit"    → ✎ editing
	// mode="preview" → 👁 previewing
	// mode="other"   → … idle
}

// Numeric comparison.
// IfOrd supports Gt, Lt, Gte, Lte for any ordered type (int, float64, string).
func ExampleOrdCondition() {
	// example:
	count := 5
	_ = IfOrd(&count).Gte(10).Then(Text("many")).Else(Text("few"))
	// :example

	c1, c2 := 5, 15
	renderAndPrint("OrdCondition", VBox(
		HBox.Gap(1)(Text("count=5  →").Dim(), IfOrd(&c1).Gte(10).Then(Text("many (≥10)").FG(Green)).Else(Text("few (<10)").FG(Yellow))),
		HBox.Gap(1)(Text("count=15 →").Dim(), IfOrd(&c2).Gte(10).Then(Text("many (≥10)").FG(Green)).Else(Text("few (<10)").FG(Yellow))),
	), 34, 2)
	// Output:
	// count=5  → few (<10)
	// count=15 → many (≥10)
}

// Slice iteration.
// Iterate a slice with per-item templates. The callback receives a pointer to each item.
func ExampleForEach() {
	// example:
	type Todo struct {
		Title string
		Done  bool
	}
	items := []Todo{{Title: "Ship it"}, {Title: "Test it"}}

	ForEach(&items, func(item *Todo) any {
		return HBox.Gap(2)(
			Checkbox(&item.Done, ""),
			Text(&item.Title),
		)
	})
}

// Inline spans.
// Mix Bold, Dim, Italic, FG and plain strings in a single line.
func ExampleRich() {
	// example:
	tree := Rich(
		Bold("Important: "),
		Dim("supporting detail"),
		FG("error", Red),
	)
	// :example

	renderAndPrint("Rich", tree, 40, 1)
	// Output: Important: supporting detailerror
}

// Custom span styles.
// Styled creates a span with a full Style struct for fine-grained control.
func ExampleSpan() {
	// example:
	tree := Rich(
		Styled("custom", Style{FG: Hex(0xFF5500), Attr: AttrBold}),
		" mixed with ",
		Italic("emphasis"),
	)
	// :example

	renderAndPrint("Span", tree, 40, 1)
	// Output: custom mixed with emphasis
}

// Full-screen app.
// SetView compiles the tree once, Handle binds keys, Run blocks until Stop is called.
func ExampleNewApp() {
	app := NewApp()

	counter := 0
	app.SetView(
		VBox(
			Text(&counter),
			Text("press q to quit"),
		),
	)
	app.Handle("q", func() { app.Stop() })
	app.Run()
}

// Inline app.
// Renders within the terminal flow instead of taking over the screen. Height sets the visible rows.
func ExampleNewInlineApp() {
	app := NewInlineApp()

	status := "loading..."
	app.SetView(Text(&status))
	app.Height(3).ClearOnExit(true)
	app.Run()
}

// Multi-view routing.
// Named views with per-view key bindings. Go switches between views; each view is compiled independently.
func ExampleApp_multiView() {
	app := NewApp()

	app.View("home", VBox(
		Text("Welcome"),
		Text("press n for next"),
	)).Handle("n", func() { app.Go("detail") })

	app.View("detail", VBox(
		Text("Detail view"),
		Text("press b for back"),
	)).Handle("b", func() { app.Go("home") })

	app.RunFrom("home")
}

// Goroutine updates.
// Mutate the value behind the pointer from any goroutine, then call RequestRender to trigger a redraw.
func ExampleApp_goroutine() {
	app := NewApp()

	status := "waiting..."
	app.SetView(Text(&status))
	app.Handle("q", func() { app.Stop() })

	time.AfterFunc(2*time.Second, func() {
		status = "done!"
		app.RequestRender()
	})

	app.Run()
}

// Views with text input.
// NoCounts disables vim-style count prefixes, preventing number keys from being swallowed before reaching an input.
func ExampleViewBuilder() {
	app := NewApp()

	var name string
	app.View("editor",
		VBox(
			Text("Editor"),
			Input().Placeholder("type here..."),
		),
	).NoCounts().Handle("<C-s>", func() {
		name = name + " saved"
	}).Handle("escape", func() {
		app.Go("home")
	})
}

// Style chaining.
// Build a style by chaining methods on DefaultStyle. CascadeStyle on a container applies it to all descendants.
func ExampleDefaultStyle() {
	// example:
	s := DefaultStyle().Bold().Foreground(Cyan)
	_ = VBox.CascadeStyle(&s)(
		Text("all children inherit bold cyan"),
	)
	// :example

	base := DefaultStyle().Foreground(BrightBlack)
	accent := DefaultStyle().Bold().Foreground(Cyan)
	renderAndPrint("DefaultStyle", VBox(
		VBox.CascadeStyle(&base).Border(BorderRounded).Title("base: dim")(
			Text("inherits dim"),
			Text("everywhere"),
		),
		VBox.CascadeStyle(&accent).Border(BorderRounded).Title("base: cyan bold")(
			Text("inherits bold cyan"),
			Text("everywhere"),
		),
	), 26, 8)
	// Output:
	// ╭─ base: dim ────────────╮
	// │inherits dim            │
	// │everywhere              │
	// ╰────────────────────────╯
	// ╭─ base: cyan bold ──────╮
	// │inherits bold cyan      │
	// │everywhere              │
	// ╰────────────────────────╯
}

// Struct literal.
// Construct a Style directly when you know the exact fields. Attr flags combine with bitwise OR.
func ExampleStyle() {
	// example:
	highlight := Style{FG: Yellow, Attr: AttrBold | AttrUnderline}
	_ = Text("warning").Style(highlight)
	// :example

	info := Style{FG: Cyan, Attr: AttrBold}
	warn := Style{FG: Yellow, Attr: AttrBold | AttrUnderline}
	err := Style{FG: Red, Attr: AttrBold}
	renderAndPrint("Style", VBox(
		Text("INFO  all systems go").Style(info),
		Text("WARN  disk 80% full").Style(warn),
		Text("ERROR connection lost").Style(err),
	), 24, 3)
	// Output:
	// INFO  all systems go
	// WARN  disk 80% full
	// ERROR connection lost
}

// Style margin.
// Margin is part of Style. Values are top, right, bottom, left (CSS order).
func ExampleStyle_margin() {
	// example:
	s := DefaultStyle().MarginTRBL(1, 2, 1, 2)
	tree := Text("margined text").Style(s)
	// :example

	renderAndPrint("Style_margin", tree, 20, 3)
	// Output: margined text
}

// Hex colour.
// Takes a uint32, not a string. Use Go hex literals.
func ExampleHex() {
	// example:
	_ = Text("branded").FG(Hex(0xFF5500))
	// :example

	renderAndPrint("Hex", VBox(
		Text("█ #FF5500").FG(Hex(0xFF5500)),
		Text("█ #00CC88").FG(Hex(0x00CC88)),
		Text("█ #8855FF").FG(Hex(0x8855FF)),
	), 16, 3)
	// Output:
	// █ #FF5500
	// █ #00CC88
	// █ #8855FF
}

// RGB colour.
// Precise 24-bit colour from red, green, blue components.
func ExampleRGB() {
	// example:
	_ = Text("vivid").FG(RGB(255, 85, 0))
	// :example

	renderAndPrint("RGB", HBox(
		Text("█").FG(RGB(255, 0, 0)),
		Text("█").FG(RGB(255, 85, 0)),
		Text("█").FG(RGB(255, 170, 0)),
		Text("█").FG(RGB(255, 255, 0)),
		Text("█").FG(RGB(0, 255, 0)),
		Text("█").FG(RGB(0, 170, 255)),
		Text("█").FG(RGB(85, 0, 255)),
		Text(" RGB spectrum"),
	), 22, 1)
	// Output: ███████ RGB spectrum
}

// Colour blending.
// LerpColor blends two colours. t=0 returns the first, t=1 returns the second, 0.5 is the midpoint.
func ExampleLerpColor() {
	// example:
	pct := 0.75
	bar := LerpColor(Red, Green, pct)
	_ = Text("75%").FG(bar)
	// :example

	children := make([]any, 20)
	for i := range children {
		t := float64(i) / 19.0
		children[i] = Text("█").FG(LerpColor(Red, Green, t))
	}
	renderAndPrint("LerpColor", HBox(children...), 22, 1)
	// Output: ████████████████████
}

// Terminal palette.
// BasicColor uses the terminal's 16-colour palette (0–15). These respect the user's terminal theme.
func ExampleBasicColor() {
	// example:
	tree := Text("theme-aware").FG(BasicColor(9))
	// :example

	renderAndPrint("BasicColor", tree, 20, 1)
	// Output: theme-aware
}

// Extended palette.
// PaletteColor uses the 256-colour extended palette.
func ExamplePaletteColor() {
	// example:
	tree := Text("orange-ish").FG(PaletteColor(214))
	// :example

	renderAndPrint("PaletteColor", tree, 20, 1)
	// Output: orange-ish
}

// Custom rendering.
// Escape hatch for custom rendering. Measure reports size, Render draws directly into the cell buffer.
func ExampleWidget() {
	// example:
	tree := Widget(
		func(availW int16) (w, h int16) { return availW, 1 },
		func(buf *Buffer, x, y, w, h int16) {
			for i := int16(0); i < w; i++ {
				buf.Set(int(x+i), int(y), Cell{Rune: '='})
			}
		},
	)
	// :example

	renderAndPrint("Widget", tree, 10, 1)
	// Output: ==========
}

// Streaming log.
// Stream lines from any io.Reader into a scrollable log view. AutoScroll keeps the view pinned to the bottom.
func ExampleLogC() {
	r := strings.NewReader("line 1\nline 2\nline 3\n")

	Log(r).MaxLines(1000).AutoScroll(true).BindVimNav()
}

// Stdin pipe.
// Pipe stdout from a subprocess or os.Stdin directly into a log view. The reader is consumed in a background goroutine.
func ExampleLogC_fromStdin() {
	Log(os.Stdin).MaxLines(5000).Grow(1)
}

// Filterable log.
// A log view with a built-in fuzzy search input. Lines are filtered as you type; matching is fzf-style.
func ExampleFilterLogC() {
	r := strings.NewReader("info: started\nerror: failed\ninfo: recovered\n")

	FilterLog(r).Placeholder("filter logs...").MaxLines(5000).BindVimNav()
}

// Scrollable canvas.
// LayerView displays a Layer, a virtual canvas that can be larger than the viewport.
func ExampleLayerViewC() {
	layer := NewLayer()

	LayerView(layer).Grow(1)
}

// Manual content.
// Write lines directly to a Layer for manual content control. The Render callback fires when the viewport resizes.
func ExampleLayer() {
	layer := NewLayer()
	layer.Render = func() {
		layer.SetLineString(0, "redrawn", DefaultStyle())
	}
	layer.SetLineString(0, "first line", DefaultStyle())
	layer.SetLineString(1, "second line", DefaultStyle().Bold())

	LayerView(layer).Grow(1)
}

// Jump target.
// Wrap any component as a jump target. When jump mode is active, labelled hints appear over each target.
func ExampleJumpC() {
	selected := "none"
	Jump(Text("clickable item"), func() {
		selected = "clicked"
		_ = selected
	})
}

// App-level jump mode.
// Press the jump key to enter jump mode, then type a target label to select it.
func ExampleJumpC_app() {
	app := NewApp()

	items := []string{"Save", "Load", "Quit"}
	app.SetView(VBox(
		Jump(Text(&items[0]), func() {}),
		Jump(Text(&items[1]), func() {}),
		Jump(Text(&items[2]), func() { app.Stop() }),
	))

	app.JumpKey("f")
	app.Run()
}

// Compile a template.
// Build compiles a UI tree into a Template. This is the same step that SetView performs internally. Useful for rendering into a Layer or for headless testing.
func ExampleBuild() {
	// example:
	tmpl := Build(VBox(
		Text("hello"),
		Text("world"),
	))
	// :example

	buf := NewBuffer(20, 2)
	tmpl.Execute(buf, 20, 2)
	fmt.Println(strings.TrimSpace(buf.StringTrimmed()))
	// Output:
	// hello
	// world
}

// Predefined theme.
// CascadeStyle sets the base style for all descendants; individual elements can override with per-element styles.
func ExampleThemeEx() {
	// example:
	theme := ThemeDark
	tree := VBox.CascadeStyle(&theme.Base).Border(BorderRounded).BorderFG(theme.Border.FG)(
		Text("normal text"),
		Text("muted").Style(theme.Muted),
		Text("accent").Style(theme.Accent),
		Text("error!").Style(theme.Error),
	)
	// :example

	renderAndPrint("ThemeEx", tree, 20, 6)
	// Output:
	// ╭──────────────────╮
	// │normal text       │
	// │muted             │
	// │accent            │
	// │error!            │
	// ╰──────────────────╯
}

// Built-in borders.
// Choose from BorderSingle, BorderRounded, or BorderDouble.
func ExampleBorderStyle() {
	// example:
	tree := VBox.Border(BorderDouble).BorderFG(Cyan)(
		Text("double-bordered"),
	)
	// :example

	renderAndPrint("BorderStyle", tree, 20, 3)
	// Output:
	// ╔══════════════════╗
	// ║double-bordered   ║
	// ╚══════════════════╝
}

// Custom borders.
// Define custom borders with any rune for each edge and corner.
func ExampleBorderStyle_custom() {
	// example:
	ascii := BorderStyle{
		Horizontal:  '-',
		Vertical:    '|',
		TopLeft:     '+',
		TopRight:    '+',
		BottomLeft:  '+',
		BottomRight: '+',
	}
	tree := VBox.Border(ascii)(Text("ascii box"))
	// :example

	renderAndPrint("BorderStyle_custom", tree, 20, 3)
	// Output:
	// +------------------+
	// |ascii box         |
	// +------------------+
}

// Text alignment.
// Align is set via the Style struct. AlignLeft, AlignCenter, or AlignRight within a fixed width.
func ExampleAlign() {
	// example:
	tree := VBox(
		Text("left-aligned"),
		Text("centered").Style(Style{Align: AlignCenter}).Width(40),
		Text("right-aligned").Style(Style{Align: AlignRight}).Width(40),
	)
	// :example

	renderAndPrint("Align", tree, 40, 3)
	// Output:
	// left-aligned
	//                 centered
	//                            right-aligned
}

// Tab shapes.
// Three tab styles: TabsStyleUnderline (default), TabsStyleBox, and TabsStyleBracket.
func ExampleTabsStyle() {
	// example:
	var active int
	tabs := []string{"Files", "Search", "Git"}
	tree := Tabs(tabs, &active).Kind(TabsStyleBracket)
	// :example

	renderAndPrint("TabsStyle", tree, 40, 1)
	// Output: [Files]  [Search]  [Git]
}

// Custom layout.
// Arrange lets you define a fully custom layout function. It receives child sizes and available space, returns rects.
func ExampleArrange() {
	// example:
	grid := Arrange(func(children []ChildSize, w, h int) []Rect {
		cols := 3
		cellW := w / cols
		rects := make([]Rect, len(children))
		for i := range children {
			rects[i] = Rect{
				X: (i % cols) * cellW,
				Y: (i / cols) * 2,
				W: cellW,
				H: 2,
			}
		}
		return rects
	})
	tree := grid(Text("a"), Text("b"), Text("c"))
	// :example

	renderAndPrint("Arrange", tree, 30, 2)
	// Output: a         b         c
}

// Scoped helpers.
// Define scopes local helper functions inside the view tree. The function runs at build time and returns a component.
func ExampleDefine() {
	// example:
	a, b, c := true, false, true
	tree := Define(func() any {
		dot := func(v *bool) any {
			return If(v).Then(Text("●").FG(Green)).Else(Text("○").FG(Red))
		}
		return HBox.Gap(1)(dot(&a), dot(&b), dot(&c))
	})
	// :example

	renderAndPrint("Define", tree, 10, 1)
	// Output: ● ○ ●
}

// Focus cycling.
// FocusManager coordinates focus between multiple inputs. Tab/Shift-Tab cycles through ManagedBy components.
func ExampleFocusManager() {
	fm := NewFocusManager()

	VBox(
		Input().Placeholder("Name").ManagedBy(fm),
		Input().Placeholder("Email").ManagedBy(fm),
	)
}

// Collapsible tree.
// TreeView renders a collapsible tree from TreeNode structs. Set Expanded to control which nodes start open.
func ExampleTreeView() {
	root := &TreeNode{
		Label: "src",
		Children: []*TreeNode{
			{Label: "main.go"},
			{Label: "lib", Children: []*TreeNode{
				{Label: "utils.go"},
				{Label: "types.go"},
			}},
		},
		Expanded: true,
	}

	_ = TreeView{Root: root, ShowRoot: true, Indent: 2, ShowLines: true}
}

// Case transform.
// Transform text case via the Style struct. Transforms run at render time without modifying the source string.
func ExampleTextTransform() {
	// example:
	tree := Text("hello world").Style(Style{Transform: TransformUppercase})
	// :example

	renderAndPrint("TextTransform", tree, 20, 1)
	// Output: HELLO WORLD
}

// Registration form.
// Form arranges labeled fields with aligned labels and automatic focus management. Mix inputs, checkboxes and validators.
func ExampleFormFn() {
	var name, email, pass string
	var terms bool

	Form.LabelBold().Gap(1)(
		Field("Name", Input(&name).Placeholder("Jane Doe")),
		Field("Email", Input(&email).Placeholder("jane@example.com")),
		Field("Password", Input(&pass).Placeholder("min 8 chars").Mask('*')),
		Field("Terms", Checkbox(&terms, "I agree")),
	)
}

// Submit handler.
// Forward-declare the form variable so OnSubmit can reference the same instance.
func ExampleFormFn_onSubmit() {
	var name, email, pass string
	var terms bool
	var form *FormC

	form = Form.LabelBold().OnSubmit(func() {
		_ = form.ValidateAll()
	})(
		Field("Name", Input(&name).Validate(VRequired)),
		Field("Email", Input(&email).Validate(VEmail)),
		Field("Password", Input(&pass).Mask('*').Validate(VMinLen(8))),
		Field("Terms", Checkbox(&terms, "I agree").Validate(VTrue)),
	)
	_ = form
}

// Mixed control types.
// Inputs, checkboxes and radio groups all work as form fields with automatic focus cycling.
func ExampleFormFn_mixedControls() {
	var name string
	var notify bool
	var role int

	Form.LabelBold().Gap(1)(
		Field("Name", Input(&name).Placeholder("you")),
		Field("Notify", Checkbox(&notify, "Send me updates")),
		Field("Role", Radio(&role, "Viewer", "Editor", "Admin")),
	)
}

// ---------------------------------------------------------------------------
// Screen effects
// ---------------------------------------------------------------------------

// Single post-processing effect.
// Place ScreenEffect anywhere in the view tree — it applies to the whole screen.
func ExampleScreenEffect() {
	ScreenEffect(SEVignette())
}

// Stacked effects.
// Effects apply in order, left to right — each sees the output of the previous.
func ExampleScreenEffect_multiple() {
	ScreenEffect(
		SEDesaturate().Strength(0.5),
		SETint(Hex(0xFF6600)).Strength(0.1),
	)
}

// Conditional effect.
// Wrap in If() to toggle reactively based on state.
func ExampleScreenEffect_conditional() {
	dimmed := false
	If(&dimmed).Then(ScreenEffect(SEDimAll()))
}

// Per-cell transform.
// The fragment shader equivalent — define per-cell logic, iteration is handled.
func ExampleEachCell() {
	EachCell(func(x, y int, c Cell, ctx PostContext) Cell {
		if y%2 == 0 {
			c.Style.Attr = c.Style.Attr.With(AttrDim)
		}
		return c
	})
}

// Blend mode wrapper.
// Snapshots the buffer, runs the effect, then blends the result back using a Photoshop-style mode.
func ExampleWithBlend() {
	ScreenEffect(WithBlend(BlendScreen, SEBloom()))
}

// Quantize wrapper.
// Snap colours to step-size buckets after an effect runs. Use step=32 to cut output size.
func ExampleWithQuantize() {
	ScreenEffect(WithQuantize(32, SEBloom()))
}

// Manual colour blend.
// BlendColor combines two colours with a Photoshop-style mode — useful inside custom effects.
func ExampleBlendColor() {
	base := RGB(100, 150, 200)
	top := RGB(255, 100, 50)
	_ = BlendColor(base, top, BlendScreen)
}

// Dim entire screen uniformly.
func ExampleSEDimAll() {
	ScreenEffect(SEDimAll())
}

// Warm colour grade.
func ExampleSETint() {
	ScreenEffect(SETint(Hex(0xFF6600)).Strength(0.15))
}

// Tint that spares a focused panel.
func ExampleSETint_dodge() {
	var panel NodeRef
	ScreenEffect(SETint(Hex(0x0066FF)).Strength(0.3).Dodge(&panel))
}

// Cinematic edge darkening.
func ExampleSEVignette() {
	ScreenEffect(SEVignette().Strength(0.7))
}

// Vignette centred on a panel.
func ExampleSEVignette_focus() {
	var panel NodeRef
	ScreenEffect(SEVignette().Focus(&panel))
}

// Vignette that skips a panel.
func ExampleSEVignette_dodge() {
	var panel NodeRef
	ScreenEffect(SEVignette().Dodge(&panel))
}

// Smooth vignette without quantization.
func ExampleSEVignette_smooth() {
	ScreenEffect(SEVignette().Smooth().Strength(0.6))
}

// Wash out colour.
func ExampleSEDesaturate() {
	ScreenEffect(SEDesaturate().Strength(0.8))
}

// Colour spotlight — grey world, one panel in colour.
func ExampleSEDesaturate_dodge() {
	var panel NodeRef
	ScreenEffect(SEDesaturate().Dodge(&panel))
}

// Punch up contrast.
func ExampleSEContrast() {
	ScreenEffect(SEContrast().Strength(2.0))
}

// Contrast that spares a panel.
func ExampleSEContrast_dodge() {
	var panel NodeRef
	ScreenEffect(SEContrast().Dodge(&panel))
}

// Dim everything outside a node.
func ExampleSEFocusDim() {
	var panel NodeRef
	ScreenEffect(SEFocusDim(&panel))
}

// Remap colours through a three-stop gradient.
func ExampleSEGradientMap() {
	ScreenEffect(SEGradientMap(
		RGB(0, 0, 50),      // shadows → deep blue
		RGB(0, 128, 128),   // midtones → teal
		RGB(200, 255, 200), // highlights → mint
	))
}

// Directional drop shadow behind a panel.
func ExampleSEDropShadow() {
	var panel NodeRef
	ScreenEffect(SEDropShadow().Focus(&panel))
}

// Symmetric glow — offset(0,0) centres the shadow source on the panel.
func ExampleSEDropShadow_glow() {
	var panel NodeRef
	ScreenEffect(SEDropShadow().Focus(&panel).Offset(0, 0).Strength(0.4).Radius(12))
}

// Colour-sampling glow that reads the panel's edge colours and spills them outward.
func ExampleSEGlow() {
	var panel NodeRef
	ScreenEffect(SEGlow().Focus(&panel))
}

// Bloom around bright cells.
func ExampleSEBloom() {
	ScreenEffect(SEBloom().Threshold(0.6).Strength(0.3))
}

// Bloom constrained to a panel.
func ExampleSEBloom_focus() {
	var panel NodeRef
	ScreenEffect(SEBloom().Focus(&panel))
}

// Green phosphor monochrome.
func ExampleSEMonochrome() {
	ScreenEffect(SEMonochrome(RGB(0, 255, 80)))
}

// Monochrome with a colour spotlight.
func ExampleSEMonochrome_dodge() {
	var panel NodeRef
	ScreenEffect(SEMonochrome(RGB(0, 255, 80)).Dodge(&panel))
}

// Snap colours to 32-level steps.
// Reduces escape output for animated effects with negligible visible banding.
func ExampleSEQuantize() {
	ScreenEffect(SEQuantize(32))
}

// First-match-wins conditional.
// Evaluates cases top-to-bottom; the first true case renders.
func ExampleMatch() {
	// example:
	status := "error"
	_ = Match(&status,
		Eq("ok", Text("all clear")),
		Eq("warn", Text("warning")),
		Eq("error", Text("failure")),
	).Default(Text("unknown"))
	// :example

	s1, s2, s3 := "ok", "warn", "error"
	row := func(s *string) any {
		return HBox.Gap(1)(Text(*s+"  →").Dim(), Match(s,
			Eq("ok", Text("✓ all clear").FG(Green)),
			Eq("warn", Text("⚠ warning").FG(Yellow)),
			Eq("error", Text("✗ failure").FG(Red)),
		).Default(Text("? unknown").Dim()))
	}
	renderAndPrint("Match", VBox(row(&s1), row(&s2), row(&s3)), 24, 3)
	// Output:
	// ok  → ✓ all clear
	// warn  → ⚠ warning
	// error  → ✗ failure
}

// Ordered thresholds.
// Order matters — the first matching case wins, so check the highest threshold first.
func ExampleMatch_ordered() {
	// example:
	cpu := 85.0
	_ = Match(&cpu,
		Gt(90.0, Text("CRITICAL")),
		Gt(70.0, Text("WARNING")),
		Lte(70.0, Text("OK")),
	)
	// :example

	c1, c2, c3 := 95.0, 75.0, 40.0
	bar := func(label string, c *float64) any {
		return HBox.Gap(1)(Text(label+" →").Dim(), Match(c,
			Gt(90.0, Text("CRITICAL").FG(Red).Bold()),
			Gt(70.0, Text("WARNING").FG(Yellow)),
			Lte(70.0, Text("OK").FG(Green)),
		))
	}
	renderAndPrint("Match_ordered", VBox(bar(" 95", &c1), bar(" 75", &c2), bar(" 40", &c3)), 24, 3)
	// Output:
	//  95 → CRITICAL
	//  75 → WARNING
	//  40 → OK
}

// Predicate matching.
// Where accepts a function for cases that need custom logic.
func ExampleMatch_where() {
	// example:
	query := "hello world"
	_ = Match(&query,
		Eq("", Text("type to search")),
		Where(func(q string) bool { return len(q) < 3 }, Text("keep typing...")),
		Where(func(q string) bool { return len(q) >= 3 }, Text("searching")),
	)
	// :example

	q1, q2, q3 := "", "hi", "hello world"
	row := func(q *string) any {
		label := "\"" + *q + "\""
		return HBox.Gap(1)(Text(label).Dim().Width(14), Match(q,
			Eq("", Text("type to search").FG(BrightBlack)),
			Where(func(s string) bool { return len(s) < 3 }, Text("keep typing...").FG(Yellow)),
			Where(func(s string) bool { return len(s) >= 3 }, Text("searching").FG(Green)),
		))
	}
	renderAndPrint("Match_where", VBox(row(&q1), row(&q2), row(&q3)), 34, 3)
	// Output:
	// ""             type to search
	// "hi"           keep typing...
	// "hello world"  searching
}
