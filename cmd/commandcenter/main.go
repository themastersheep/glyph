// commandcenter: dense live dashboard demo with service drill-down
package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	. "github.com/kungfusheep/glyph"
	"github.com/kungfusheep/riffkey"
)

type palette struct {
	base    Color
	surface Color
	overlay Color
	text    Color
	subtle  Color
	muted   Color
	love    Color // error/warn accent
	gold    Color // warning/highlight
	foam    Color // success/healthy
	iris    Color // info/accent
}

var themes = []struct {
	name string
	pal  palette
}{
	{"rose piné", palette{
		base: Hex(0x191724), surface: Hex(0x1f1d2e), overlay: Hex(0x26233a),
		text: Hex(0xe0def4), subtle: Hex(0x908caa), muted: Hex(0x6e6a86),
		love: Hex(0xeb6f92), gold: Hex(0xf6c177), foam: Hex(0x9ccfd8), iris: Hex(0xc4a7e7),
	}},
	{"catppuccin mocha", palette{
		base: Hex(0x1e1e2e), surface: Hex(0x313244), overlay: Hex(0x45475a),
		text: Hex(0xcdd6f4), subtle: Hex(0xa6adc8), muted: Hex(0x6c7086),
		love: Hex(0xf38ba8), gold: Hex(0xf9e2af), foam: Hex(0xa6e3a1), iris: Hex(0xcba6f7),
	}},
	{"mfd", palette{
		base: Hex(0x7A8B69), surface: Hex(0x5A6B4A), overlay: Hex(0x5A6B4A),
		text: Hex(0x1E2D1E), subtle: Hex(0x3A4A3A), muted: Hex(0x354828),
		love: Hex(0x0D1D0D), gold: Hex(0x0D1D0D), foam: Hex(0x1E2D1E), iris: Hex(0x1E2D1E),
	}},
	{"gruvbox", palette{
		base: Hex(0x282828), surface: Hex(0x3c3836), overlay: Hex(0x504945),
		text: Hex(0xebdbb2), subtle: Hex(0xa89984), muted: Hex(0x665c54),
		love: Hex(0xfb4934), gold: Hex(0xfabd2f), foam: Hex(0xb8bb26), iris: Hex(0xd3869b),
	}},
	{"nord", palette{
		base: Hex(0x2e3440), surface: Hex(0x3b4252), overlay: Hex(0x434c5e),
		text: Hex(0xeceff4), subtle: Hex(0xd8dee9), muted: Hex(0x4c566a),
		love: Hex(0xbf616a), gold: Hex(0xebcb8b), foam: Hex(0xa3be8c), iris: Hex(0xb48ead),
	}},
	{"dracula", palette{
		base: Hex(0x282a36), surface: Hex(0x44475a), overlay: Hex(0x6272a4),
		text: Hex(0xf8f8f2), subtle: Hex(0xbfbfbf), muted: Hex(0x6272a4),
		love: Hex(0xff5555), gold: Hex(0xf1fa8c), foam: Hex(0x50fa7b), iris: Hex(0xbd93f9),
	}},
	{"mfd-dark", palette{
		base: Hex(0x1E2D1E), surface: Hex(0x253525), overlay: Hex(0x3A4A3A),
		text: Hex(0x8A9B70), subtle: Hex(0x6A7B5A), muted: Hex(0x607258),
		love: Hex(0xA0B180), gold: Hex(0xA0B180), foam: Hex(0xA0B180), iris: Hex(0x8A9B70),
	}},
	{"mfd-mono", palette{
		base: Hex(0x08080C), surface: Hex(0x0C0C10), overlay: Hex(0x2A2A32),
		text: Hex(0xD0D0D8), subtle: Hex(0x909098), muted: Hex(0x606068),
		love: Hex(0xF0F0FF), gold: Hex(0xF0F0FF), foam: Hex(0xD0D0D8), iris: Hex(0xD0D0D8),
	}},
	{"mfd-amber", palette{
		base: Hex(0x0F0C08), surface: Hex(0x141008), overlay: Hex(0x3A2810),
		text: Hex(0xCC9944), subtle: Hex(0x9A7730), muted: Hex(0x7A5830),
		love: Hex(0xFFBB55), gold: Hex(0xFFBB55), foam: Hex(0xCC9944), iris: Hex(0xCC9944),
	}},
}

type service struct {
	Name       string
	Status     string
	CPU        float64
	CPUStr     string
	Mem        string
	CPUHistory []float64
}

type dashboard struct {
	rng *rand.Rand
	pal palette

	// sparkline data
	reqData, latData, errData []float64

	// metric labels
	reqRate, p99Lat, errRate, clock string

	services    []service
	selectedPtr *service
	selectedSvc service
	logR        *io.PipeReader
	logW        *io.PipeWriter

	// ui flags
	colorAnimDur  time.Duration
	showModal     bool
	sparkExpanded bool
	logExpanded   bool
	restarting    bool
	restartPct    int
	degraded      bool
	themeIdx      int

	// set after init
	app     *App
	svcList *FilterListC[service]
	logView *LogC

	// cascade styles (pointers so they update reactively with theme)
	textCascade   Style
	subtleCascade Style

	// recovery lerp
	wasDegraded  bool
	recoveryTick int
}

func newDashboard() *dashboard {
	rng := rand.New(rand.NewSource(42))
	clock := time.Now().Format("15:04:05")

	pr, pw := io.Pipe()

	s := &dashboard{
		rng:     rng,
		pal:     themes[0].pal,
		logR:    pr,
		logW:    pw,
		reqData: make([]float64, 40),
		latData: make([]float64, 40),
		errData: make([]float64, 40),
		reqRate: "52 req/s",
		p99Lat:  "58ms",
		errRate: "5.8%",
		clock:   clock,
		services: []service{
			{Name: "api-gateway", Status: "live", CPU: 7.2, CPUStr: "  7.2%", Mem: "240 MB"},
			{Name: "postgres-primary", Status: "live", CPU: 3.8, CPUStr: "  3.8%", Mem: "1.2 GB"},
			{Name: "redis-cluster", Status: "warn", CPU: 19.5, CPUStr: " 19.5%", Mem: "380 MB"},
			{Name: "worker-pool", Status: "live", CPU: 4.9, CPUStr: "  4.9%", Mem: "190 MB"},
			{Name: "cdn-edge", Status: "live", CPU: 1.1, CPUStr: "  1.1%", Mem: " 42 MB"},
			{Name: "auth-service", Status: "live", CPU: 5.4, CPUStr: "  5.4%", Mem: "128 MB"},
		},
		colorAnimDur: 5 * time.Second,
		degraded:     true,
		wasDegraded:  true,
	}

	// pre-seed sparkline history: healthy on left, degrading toward right
	for i := range s.reqData {
		blend := math.Min(1, float64(i)/float64(len(s.reqData)-1)*1.5)
		s.reqData[i] = lerp(125, 50, blend)
		s.latData[i] = lerp(22, 60, blend)
		s.errData[i] = lerp(0.8, 6, blend)
	}

	for i := range s.services {
		s.services[i].CPUHistory = make([]float64, 46)
		for j := range s.services[i].CPUHistory {
			if s.services[i].Status == "warn" {
				s.services[i].CPUHistory[j] = 15 + rng.Float64()*10
			} else {
				s.services[i].CPUHistory[j] = s.services[i].CPU + rng.Float64()*5 - 2.5
			}
		}
	}

	s.selectedSvc = s.services[0]
	s.applyTheme()

	// seed initial log lines (written async, consumed when Log component starts)
	go func() {
		fmt.Fprintf(pw, "%s  GET    /api/cache/warm     503  812ms\n", clock)
		fmt.Fprintf(pw, "%s  GET    /api/users          200   11ms\n", clock)
		fmt.Fprintf(pw, "%s  POST   /api/sessions       500  1.2s \n", clock)
	}()

	return s
}

func (s *dashboard) applyTheme() {
	s.textCascade = Style{FG: s.pal.text}
	s.subtleCascade = Style{FG: s.pal.subtle}
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func main() {
	s := newDashboard()

	var modalRouter *riffkey.Router
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	colorAnim := Animate.Duration(&s.colorAnimDur).Ease(EaseOutCubic)

	metricPanel := func(title string, data *[]float64, label *string, col any) any {
		return VBox.Grow(1).Border(BorderRounded).BorderFG(&s.pal.overlay).Title(title)(
			Sparkline(data).FG(col).Height(
				Animate.Duration(200*time.Millisecond).Ease(EaseOutCubic)(
					If(&s.sparkExpanded).Then(18).Else(1),
				),
			),
			Text(label).FG(&s.pal.muted),
		)
	}

	svcList := FilterList(&s.services, func(svc *service) string { return svc.Name }).
		Placeholder("filter...").
		Render(func(svc *service) any {
			return HBox.Gap(2)(
				VBox.Width(1)(Switch(&svc.Status).
					Case("warn", Text("○").FG(&s.pal.gold)).
					Default(Text("●").FG(&s.pal.foam))),
				Text(&svc.Name).FG(&s.pal.text),
				Space(),
				Text(&svc.CPUStr).FG(&s.pal.subtle).Width(6).Align(AlignRight),
				Text(&svc.Mem).FG(&s.pal.subtle).Width(8).Align(AlignRight),
				VBox.Width(11)(
					Switch(&svc.Status).
						Case("warn", Text("degraded").FG(&s.pal.gold)).
						Default(Text("healthy").FG(&s.pal.muted)),
				),
			)
		}).
		Handle("<Enter>", func(svc *service) {
			s.selectedPtr = svc
			s.selectedSvc = *svc
			s.showModal = true
			app.Push(modalRouter)
		}).
		HandleClear("<Esc>", nil)

	s.app = app
	s.svcList = svcList

	logView := Log(s.logR).
		MaxLines(100).
		Grow(1).
		Colorize(func(line string) []Span {
			logStyle := Style{FG: s.pal.muted, BG: s.pal.base}
			pos := 0
			for _, f := range strings.Fields(line) {
				idx := strings.Index(line[pos:], f)
				if idx < 0 {
					break
				}
				start := pos + idx
				if len(f) == 3 && f[0] >= '1' && f[0] <= '5' {
					st := logStyle
					if f[0] == '5' {
						st.FG = s.pal.love
					} else if f[0] == '2' || f[0] == '3' {
						st.FG = s.pal.foam
					}
					return []Span{
						{Text: line[:start], Style: logStyle},
						{Text: f, Style: st},
						{Text: line[start+3:], Style: logStyle},
					}
				}
				pos = start + len(f)
			}
			return []Span{{Text: line, Style: logStyle}}
		})
	s.logView = logView

	closeModal := func() {
		s.showModal = false
		s.restarting = false
		s.restartPct = 0
		app.Pop()
		app.RequestRender()
	}

	restartComplete := func() {
		s.selectedSvc.Status = "live"
		if s.selectedPtr != nil {
			s.selectedPtr.Status = "live"
			s.selectedPtr.CPU = 4.0 + s.rng.Float64()*2
		}
		closeModal()
		// shorten color anim after recovery transition plays out
		go func() {
			time.Sleep(s.colorAnimDur)
			s.colorAnimDur = 10 * time.Millisecond
		}()
	}

	anim := Animate.Duration(900 * time.Millisecond).Ease(EaseOutCubic).From(0.0)

	var popupRef NodeRef
	app.SetView(
		VBox.Grow(1).Fill(&s.pal.base).CascadeStyle(&s.textCascade)(
			VBox.Grow(1).MarginVH(1, 2)(
				HBox.CascadeStyle(&s.subtleCascade)(
					Text("● glyph control").FG(&s.pal.iris),
					Space(),
					Text("prod-us-east-1  "),
					Text(&s.clock),
				),
				HRule().Char(BorderDouble.Horizontal).FG(&s.pal.overlay),

				HBox.Gap(1)(
					metricPanel("requests/s", &s.reqData, &s.reqRate,
						colorAnim(If(&s.degraded).Then(&s.pal.gold).Else(&s.pal.foam))),
					metricPanel("p99 latency", &s.latData, &s.p99Lat,
						colorAnim(If(&s.degraded).Then(&s.pal.love).Else(&s.pal.iris))),
					metricPanel("error rate", &s.errData, &s.errRate,
						colorAnim(If(&s.degraded).Then(&s.pal.love).Else(&s.pal.gold))),
				),

				VBox.Grow(1)(
					VBox.Grow(1).Border(BorderRounded).BorderFG(&s.pal.overlay).Title("services")(
						HBox.Gap(2)(
							VBox.Width(1)(Text("●").FG(&s.pal.muted)),
							Text("SERVICE").FG(&s.pal.muted),
							Space(),
							Text("CPU").FG(&s.pal.muted).Width(6).Align(AlignRight),
							Text("MEM").FG(&s.pal.muted).Width(8),
							VBox.Width(11)(Text("STATUS").FG(&s.pal.muted)),
						),
						HRule().FG(&s.pal.overlay).Extend(),
						svcList,
					),
					VBox.Border(BorderRounded).BorderFG(&s.pal.overlay).Title("log").Height(
						Animate.Duration(200*time.Millisecond).Ease(EaseOutCubic)(
							If(&s.logExpanded).Then(16).Else(5),
						),
					)(
						logView,
					),
				),

				HRule().Char(BorderDouble.Horizontal).FG(&s.pal.overlay),
				Text("[enter] inspect  [ctrl+s] sparklines  [ctrl+l] log  [ctrl+c] quit").FG(&s.pal.muted),

				If(&s.showModal).Then(OverlayNode{
					Centered: true,
					Child: VBox.Width(46).Fill(&s.pal.surface).NodeRef(&popupRef)(
						SpaceH(1),
						HBox(
							If(&s.selectedSvc.Status).Eq("warn").
								Then(Text("  ○ ").FG(&s.pal.gold)).
								Else(Text("  ● ").FG(&s.pal.foam)),
							If(&s.selectedSvc.Status).Eq("warn").
								Then(Text(&s.selectedSvc.Name).FG(&s.pal.gold).Bold()).
								Else(Text(&s.selectedSvc.Name).FG(&s.pal.foam).Bold()),
						),
						HRule().FG(&s.pal.overlay),
						Text("  cpu history").FG(&s.pal.muted),
						Sparkline(&s.selectedSvc.CPUHistory).FG(&s.pal.iris),
						HBox.Gap(3)(
							VBox(
								Text("  cpu").FG(&s.pal.muted),
								Text("  mem").FG(&s.pal.muted),
							),
							VBox(
								IfOrd(&s.selectedSvc.CPU).Gt(20.0).
									Then(Text(&s.selectedSvc.CPUStr).FG(&s.pal.gold)).
									Else(Text(&s.selectedSvc.CPUStr).FG(&s.pal.text)),
								Text(&s.selectedSvc.Mem).FG(&s.pal.text),
							),
						),
						HRule().FG(&s.pal.overlay),
						If(&s.restarting).
							Then(
								HBox.Gap(1)(
									Text("  restarting").FG(&s.pal.muted),
									HBox.Grow(1)(
										Progress(
											Animate.Duration(2*time.Second).Ease(EaseOutCubic).OnComplete(restartComplete)(&s.restartPct),
										).FG(&s.pal.foam),
									),
								),
							).
							Else(
								Text("  [r] restart service  [esc] close").FG(&s.pal.muted),
							),
						SpaceH(1),
						ScreenEffect(
							SEVignette().Dodge(&popupRef).Smooth().Strength(anim(0.88)),
							SEGlow().Focus(&popupRef).Brightness(1.1).Strength(anim(0.5)),
						),
					),
				}),
			),
		),
	)

	app.Handle("<C-c>", func() { app.Stop() })
	app.Handle("<Escape>", func() { s.svcList.Clear() })
	app.Handle("<C-s>", func() {
		s.sparkExpanded = !s.sparkExpanded
		if s.logExpanded {
			s.logExpanded = false
		}
	})
	app.Handle("<C-l>", func() {
		s.logExpanded = !s.logExpanded
		if s.sparkExpanded {
			s.sparkExpanded = false
		}
	})
	app.Handle("<C-t>", func() {
		s.themeIdx = (s.themeIdx + 1) % len(themes)
		s.pal = themes[s.themeIdx].pal
		s.applyTheme()
		s.logView.Refresh()
	})

	modalRouter = riffkey.NewRouter()
	modalRouter.Handle("<Escape>", func(_ riffkey.Match) { closeModal() })
	modalRouter.Handle("<C-c>", func(_ riffkey.Match) { app.Stop() })
	modalRouter.Handle("r", func(_ riffkey.Match) {
		if s.restarting {
			return
		}
		s.restarting = true
		s.restartPct = 100
	})

	go func() {
		for range time.NewTicker(400 * time.Millisecond).C {
			s.tick()
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

const recoveryTicks = 10

func (s *dashboard) tick() {
	hasDegraded := false
	for _, svc := range s.services {
		if svc.Status == "warn" {
			hasDegraded = true
			break
		}
	}

	// detect degraded→healthy transition
	if s.wasDegraded && !hasDegraded {
		s.recoveryTick = 0
	}
	s.wasDegraded = hasDegraded

	var rps, lat, er float64
	if hasDegraded {
		rps = 50 + s.rng.Float64()*2
		lat = 60 + s.rng.Float64()*2
		er = 6 + s.rng.Float64()*0.3
	} else if s.recoveryTick < recoveryTicks {
		s.recoveryTick++
		t := float64(s.recoveryTick) / float64(recoveryTicks)
		t = 1 - math.Pow(1-t, 2)
		rps = lerp(50, 125, t) + s.rng.Float64()*2
		lat = lerp(60, 22, t) + s.rng.Float64()*2
		er = lerp(6, 0.8, t) + s.rng.Float64()*0.3
	} else {
		rps = 125 + s.rng.Float64()*2
		lat = 22 + s.rng.Float64()*2
		er = 0.8 + s.rng.Float64()*0.3
	}

	copy(s.reqData, s.reqData[1:])
	s.reqData[len(s.reqData)-1] = rps
	copy(s.latData, s.latData[1:])
	s.latData[len(s.latData)-1] = lat
	copy(s.errData, s.errData[1:])
	s.errData[len(s.errData)-1] = er

	s.reqRate = fmt.Sprintf("%.0f req/s", rps)
	s.p99Lat = fmt.Sprintf("%.0fms", lat)
	s.errRate = fmt.Sprintf("%.1f%%", er)
	s.clock = time.Now().Format("15:04:05")

	s.degraded = hasDegraded

	for i := range s.services {
		if s.services[i].Name == "redis-cluster" && s.services[i].Status == "warn" {
			s.services[i].CPU = math.Max(15, math.Min(25, s.services[i].CPU+s.rng.Float64()*4-2))
		} else {
			s.services[i].CPU = math.Max(0.5, s.services[i].CPU+s.rng.Float64()*2-1)
		}
		s.services[i].CPUStr = fmt.Sprintf("%5.1f%%", s.services[i].CPU)
		copy(s.services[i].CPUHistory, s.services[i].CPUHistory[1:])
		s.services[i].CPUHistory[len(s.services[i].CPUHistory)-1] = s.services[i].CPU
	}
	s.svcList.Refresh()

	if hasDegraded {
		fmt.Fprintf(s.logW, "%s  %-6s %-22s %d  %s\n",
			s.clock,
			[]string{"GET", "GET", "POST", "GET", "PUT"}[s.rng.Intn(5)],
			[]string{"/api/cache/warm", "/api/sessions", "/api/users", "/api/health", "/api/orders"}[s.rng.Intn(5)],
			[]int{503, 500, 502, 200, 200, 503, 500, 200}[s.rng.Intn(8)],
			[]string{"340ms", "812ms", "1.1s ", "1.4s ", "22ms", "680ms"}[s.rng.Intn(6)],
		)
	} else {
		fmt.Fprintf(s.logW, "%s  %-6s %-22s %d  %dms\n",
			s.clock,
			[]string{"GET", "POST", "GET", "PUT", "DELETE"}[s.rng.Intn(5)],
			[]string{"/api/users", "/api/deploy", "/api/health", "/api/orders", "/api/metrics"}[s.rng.Intn(5)],
			[]int{200, 200, 200, 201, 204, 200, 200, 304}[s.rng.Intn(8)],
			2+s.rng.Intn(45),
		)
	}

	s.app.RequestRender()
}
