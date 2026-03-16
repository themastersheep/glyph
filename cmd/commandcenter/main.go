// commandcenter: dense live dashboard demo with service drill-down
package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"time"

	. "github.com/kungfusheep/glyph"
	"github.com/kungfusheep/riffkey"
)

// Rose Piné palette
var (
	rpBase    = Hex(0x191724)
	rpSurface = Hex(0x1f1d2e)
	rpOverlay = Hex(0x26233a)
	rpText    = Hex(0xe0def4)
	rpSubtle  = Hex(0x908caa)
	rpMuted   = Hex(0x6e6a86)
	rpLove    = Hex(0xeb6f92)
	rpGold    = Hex(0xf6c177)
	rpFoam    = Hex(0x9ccfd8)
	rpIris    = Hex(0xc4a7e7)
)

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
	showModal     bool
	sparkExpanded bool
	logExpanded   bool
	restarting    bool
	restartPct    int
	degraded      bool
	pulseStyle    Style

	// set after init
	app     *App
	svcList *FilterListC[service]

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
		degraded:    true,
		wasDegraded: true,
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

	// seed initial log lines (written async, consumed when Log component starts)
	go func() {
		fmt.Fprintf(pw, "%s  GET    /api/cache/warm     503  812ms\n", clock)
		fmt.Fprintf(pw, "%s  GET    /api/users          200   11ms\n", clock)
		fmt.Fprintf(pw, "%s  POST   /api/sessions       500  1.2s \n", clock)
	}()

	return s
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func main() {
	s := newDashboard()

	var modalRouter *riffkey.Router
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	colorAnim := Animate.Duration(5 * time.Second).Ease(EaseOutCubic)

	metricPanel := func(title string, data *[]float64, label *string, col any) any {
		return VBox.Grow(1).Border(BorderRounded).BorderFG(rpOverlay).Title(title)(
			Sparkline(data).FG(col).Height(
				Animate.Duration(200*time.Millisecond).Ease(EaseOutCubic)(
					If(&s.sparkExpanded).Then(18).Else(1),
				),
			),
			Text(label).FG(rpMuted),
		)
	}

	svcList := FilterList(&s.services, func(svc *service) string { return svc.Name }).
		Placeholder("filter...").
		Render(func(svc *service) any {
			return HBox.Gap(2)(
				VBox.Width(1)(Switch(&svc.Status).
					Case("warn", Text("○").FG(rpGold)).
					Default(Text("●").FG(rpFoam))),
				Text(&svc.Name).FG(rpText),
				Space(),
				Text(&svc.CPUStr).FG(rpSubtle).Width(6).Align(AlignRight),
				Text(&svc.Mem).FG(rpSubtle).Width(8).Align(AlignRight),
				VBox.Width(11)(
					Switch(&svc.Status).
						Case("warn", Text("degraded").FG(rpGold)).
						Default(Text("healthy").FG(rpMuted)),
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
		FG(rpMuted).
		BG(rpBase).
		OnUpdate(app.RequestRender)

	anim := Animate.Duration(900 * time.Millisecond).Ease(EaseOutCubic).From(0.0)

	var popupRef NodeRef
	app.SetView(
		VBox.Grow(1).Fill(rpBase).CascadeStyle(&Style{FG: rpText})(
			VBox.Grow(1).MarginVH(1, 2)(
				HBox.CascadeStyle(&Style{FG: rpSubtle})(
					Text("● glyph control").FG(rpIris),
					Space(),
					Text("prod-us-east-1  "),
					Text(&s.clock),
				),
				HRule().Char(BorderDouble.Horizontal).FG(rpOverlay),

				HBox.Gap(1)(
					metricPanel("requests/s", &s.reqData, &s.reqRate,
						colorAnim(If(&s.degraded).Then(rpGold).Else(rpFoam))),
					metricPanel("p99 latency", &s.latData, &s.p99Lat,
						colorAnim(If(&s.degraded).Then(rpLove).Else(rpIris))),
					metricPanel("error rate", &s.errData, &s.errRate,
						colorAnim(If(&s.degraded).Then(rpLove).Else(rpGold))),
				),

				VBox.Grow(1)(
					VBox.Grow(1).Border(BorderRounded).BorderFG(rpOverlay).Title("services")(
						HBox.Gap(2)(
							Text("●  SERVICE").FG(rpMuted),
							Space(),
							Text("CPU").FG(rpMuted).Width(5).Align(AlignRight),
							Text("MEM").FG(rpMuted).Width(7).Align(AlignRight),
							Text("STATUS").FG(rpMuted).Width(13),
						),
						HRule().FG(rpOverlay).Extend(),
						svcList,
					),
					VBox.Border(BorderRounded).BorderFG(rpOverlay).Title("log").Height(
						Animate.Duration(200*time.Millisecond).Ease(EaseOutCubic)(
							If(&s.logExpanded).Then(16).Else(5),
						),
					)(
						logView,
					),
				),

				HRule().Char(BorderDouble.Horizontal).FG(rpOverlay),
				Text("[enter] inspect  [ctrl+s] sparklines  [ctrl+l] log  [ctrl+c] quit").FG(rpMuted),

				If(&s.showModal).Then(OverlayNode{
					Centered: true,
					Child: VBox.Width(46).Fill(rpSurface).NodeRef(&popupRef)(
						SpaceH(1),
						HBox(
							If(&s.selectedSvc.Status).Eq("warn").
								Then(Text("  ○ ").FG(rpGold)).
								Else(Text("  ● ").FG(rpFoam)),
							If(&s.selectedSvc.Status).Eq("warn").
								Then(Text(&s.selectedSvc.Name).FG(rpGold).Bold()).
								Else(Text(&s.selectedSvc.Name).FG(rpFoam).Bold()),
						),
						HRule().FG(rpOverlay),
						Text("  cpu history").FG(rpMuted),
						Sparkline(&s.selectedSvc.CPUHistory).FG(rpIris),
						HBox.Gap(3)(
							VBox(
								Text("  cpu").FG(rpMuted),
								Text("  mem").FG(rpMuted),
							),
							VBox(
								IfOrd(&s.selectedSvc.CPU).Gt(20.0).
									Then(Text(&s.selectedSvc.CPUStr).FG(rpGold)).
									Else(Text(&s.selectedSvc.CPUStr).FG(rpText)),
								Text(&s.selectedSvc.Mem).FG(rpText),
							),
						),
						HRule().FG(rpOverlay),
						If(&s.restarting).
							Then(
								HBox.Gap(1)(
									Text("  restarting").FG(rpMuted),
									HBox.Grow(1).CascadeStyle(&s.pulseStyle)(
										Progress(&s.restartPct),
									),
								),
							).
							Else(
								Text("  [r] restart service  [esc] close").FG(rpMuted),
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
	app.Handle("<Escape>", func() {})
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

	closeModal := func() {
		s.showModal = false
		s.restarting = false
		app.Pop()
		app.RequestRender()
	}

	modalRouter = riffkey.NewRouter()
	modalRouter.Handle("<Escape>", func(_ riffkey.Match) { closeModal() })
	modalRouter.Handle("<C-c>", func(_ riffkey.Match) { app.Stop() })
	modalRouter.Handle("r", func(_ riffkey.Match) {
		if s.restarting {
			return
		}
		s.restarting = true
		s.restartPct = 0
		go func() {
			for i := 1; i <= 50; i++ {
				time.Sleep(40 * time.Millisecond)
				s.restartPct = i * 2
				t := float64(s.restartPct) / 100.0
				b := math.Pow(t, 1.5)
				s.pulseStyle.FG = RGB(uint8(30+b*80), uint8(120+b*100), uint8(110+b*90))
				app.RequestRender()
			}
			s.selectedSvc.Status = "live"
			if s.selectedPtr != nil {
				s.selectedPtr.Status = "live"
				s.selectedPtr.CPU = 4.0 + s.rng.Float64()*2
			}
			closeModal()
		}()
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
