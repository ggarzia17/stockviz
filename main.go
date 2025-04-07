package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var reset = "\033[0m"
var bold = "\033[1m"
var underline = "\033[4m"


// var cRed = "\033[31m"
// var cLightRed = "\033[91m"
// var cGreen = "\033[32m"
// var cLightGreen = "\033[92m"
// var cGrey = "\033[90m"
var cBlack = "\033[30m"
var cWhite = "\033[97m"


var cRedBg = "\033[41m"
var cLightRedBg = "\033[101m"
var cGreenBg = "\033[42m"
var cLightGreenBg = "\033[102m"
var cGreyBg = "\033[100m"

// types for parsing json
type Stocks struct {
	Stocks []StockInfo `json:"stocks"`
	totalCap int
}

type StockInfo struct {
	Ticker string  `json:"ticker"`
	Name string  `json:"name"`
	Sector string  `json:"sector"`
	Industry string  `json:"industry"`
	MarketCap int `json:"marketCap"`
	OpenPrice float64 `json:"open"`
	CurrentPrice float64 `json:"price"`
}

// model for ui
type model struct {
	sectors []sectormodel
	//a for all, s for sector, i for industry
	viewMode byte
	_return byte 
	currentSector sectormodel
	currentIndustry industrymodel
	textInput string
	totalMarketCap float64
	x int
	y int
	ready bool
}

type sectormodel struct {
	name string
	marketCap float64
	industries []industrymodel
}

type industrymodel struct {
	name string
	marketCap float64
	stocks []stockmodel
}

type stockmodel struct {
	ticker string
	performance float64
	marketCap float64
	color string
}

type shape struct {
	x int
	y int
	lines []string
}

func initialModel(stocks Stocks) model {
	m := model{
		sectors: []sectormodel{},
		textInput: "",
		viewMode: 'a',
		currentSector: sectormodel{},
		currentIndustry: industrymodel{},
		totalMarketCap: 0,
		ready: false,
	}

	for i, s := range stocks.Stocks {
		p := (s.CurrentPrice - s.OpenPrice) / s.OpenPrice
		c := cRedBg
		if p > 0.02 {
			c = cGreenBg
		} else if p > 0.0075 {
			c = cLightGreenBg
		} else if p > -0.0075 {
			c = cGreyBg
		} else if p > -0.02 {
			c = cLightRedBg
		}

		newStock := stockmodel{
			ticker: s.Ticker,
			marketCap: float64(s.MarketCap),
			performance: p,
			color: c,
		}

		newIndustry := industrymodel{
			name: s.Industry,
			stocks: []stockmodel{newStock},
		}

		found := false
		for j, sec := range m.sectors {
			if sec.name == stocks.Stocks[i].Sector {
				for k, ind := range sec.industries {
					if ind.name == stocks.Stocks[i].Industry {
						m.sectors[j].industries[k].stocks = append(m.sectors[j].industries[k].stocks, newStock) 
						found = true
					}
				}
				if !found {
					m.sectors[j].industries = append(m.sectors[j].industries, newIndustry)
				}
				found = true
			}
		}
		if !found {
			m.sectors = append(m.sectors, sectormodel{name: s.Sector, industries: []industrymodel{newIndustry}})
		}
	}

	for i, sec := range m.sectors {
		secCap := 0.0
		for j, ind := range sec.industries {
			indCap := 0.0
			for _, stock := range ind.stocks {
				indCap += stock.marketCap
			}
			sec.industries[j].marketCap = indCap
			secCap += indCap
		}
		m.sectors[i].marketCap = secCap
		m.totalMarketCap += secCap
	}

	for _, sec := range m.sectors {
		for _, ind := range sec.industries {
			slices.SortFunc(ind.stocks, func(i,j stockmodel) int{
				return -cmp.Compare(i.marketCap, j.marketCap)
			})
		}
		slices.SortFunc(sec.industries, func(i,j industrymodel) int{
			return -cmp.Compare(i.marketCap, j.marketCap)
		})
	}
	slices.SortFunc(m.sectors, func(i,j sectormodel) int{
		return -cmp.Compare(i.marketCap, j.marketCap)
	})
	m.currentSector = m.sectors[10]
	// m.currentIndustry = m.sectors[0].industries[0]
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// alpha := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQUSTUVWXYZ"
	var err tea.Cmd = nil
	switch msg := msg.(type){
	case tea.KeyMsg:
		// for testing key msgs
		// m.textInput = msg.String()
		if m.viewMode == 'h'{
			switch msg.String(){
			case "ctrl+c":
				return m, tea.Quit
			case "q":
				m.viewMode = m._return
				return m, nil
			}
		}
		switch msg.String(){
		case "ctrl+c":
			return m, tea.Quit
		case "backspace":
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
		case "ctrl+h":
			m.textInput = ""
		case "enter":
			m, err = handleCommand(m)
			m.textInput = ""
		default:
			if len(msg.String()) == 1{
				m.textInput += msg.String()
			}
		}
	case tea.WindowSizeMsg:
		m.x, m.y = msg.Width, msg.Height
		m.ready = true
	}
	return m, err
}

func (m model) View() string {
	str := ""
	
	prompt := ""
	if m.ready {
		switch m.viewMode {
		case 'a':
			s := m.sectors
			
			sectors := buildSectors(m.x, m.y-2, s, m.totalMarketCap, "S&P 500")

			for _, l := range sectors.lines{
				str += l + "\n"
			}
			prompt = "Enter Sector Name> " + m.textInput + "\n"
		case 's':
			s := m.currentSector.industries
			
			industries := buildIndustries(m.x, m.y-2, s, m.currentSector.marketCap, m.currentSector.name, false)

			for _, l := range industries.lines{
				str += l + "\n"
			}
			prompt = "Enter Industry Name> " + m.textInput + "\n"
		case 'i':
			s := m.currentIndustry.stocks
			

			stocks := buildStocks(m.x, m.y-2, s, m.currentIndustry.marketCap, m.currentIndustry.name, false)
			for _, l := range stocks.lines{
				str += l + "\n"
			}
			prompt = "Enter Command > " + m.textInput + "\n"
		case 'h':
			str += "Help Menu\n"
			str += ".help, .h, ? - Shows this menu\n"
			str += ".back, .b, .. - Return to the previous screen\n"
			str += ".quit, .q, ctrl+c - Quit the program\n"
			str += "Any other text will be prefix matched with the current Sectors or Industries \nshown on screen and navigate to the largest one it finds\n"
			str += "Press q to return"
		}
	}
	str += prompt
	return str
}

func handleCommand(m model) (model, tea.Cmd) {
	switch m.textInput {
	case ".back", ".b", "..":
		switch m.viewMode {
		case 's':
			m.viewMode = 'a'
		case 'i':
			m.viewMode = 's'
		}
	case ".help", ".h", "?":
		m._return = m.viewMode
		m.viewMode = 'h'
	case ".quit", ".q":
		return m, tea.Quit
	default:
		m = prefixMatch(m)
	}
	return m, nil
}

func prefixMatch(m model) model{
	switch m.viewMode{
	case 'a':
		for _, sec := range m.sectors {
			if len(sec.name) < len(m.textInput){
				continue
			}
			if strings.EqualFold(sec.name[:len(m.textInput)], m.textInput) {
				m.viewMode = 's'
				m.currentSector = sec
				break
			}
		}
	case 's':
		for _, ind := range m.currentSector.industries {
			if strings.EqualFold(ind.name[:len(m.textInput)], m.textInput) {
				m.viewMode = 'i'
				m.currentIndustry = ind
				break
			}
		}
	}
	return m
}

// create a rectangle of width x - 0.5 and height y - 0.5
// this rectangle already has the bottom and right gaps on it hence the - 0.5
// color the rectangle cbg
// uses background colors and makes the foreground black to match the terminal (dont need to do this anymore but already wrote the code)
func rect(x, y int, cbg, t string, performance float64) shape{
	hdr := cbg + cBlack
	if x == 0 || y == 0 {
		return shape{max(x, 0),max(y, 0),[]string{}}
	}
	p := fmt.Sprint(math.Round(10000*performance)/100)+"%"
	lines := make([]string, y-1)
	for i := range lines {
		lines[i] += hdr
		if x-1 == len(t) && i == y/2-1{
			lines[i] += cWhite + t + hdr
		}else if x > len(t) && i == y/2-1{
			for range x/2-len(t)/2{
				lines[i] += " "
			}
			lines[i] += cWhite + t + hdr
			for range int(math.Ceil(float64(x)/2))-int(math.Ceil(float64(len(t))/2)) - 1{
				lines[i] += " "
			}
		}else if x > len(t) && x-1 == len(p) && i == y/2{
			lines[i] += cWhite + p + hdr
		}else if x > len(t) && x > len(p) && i == y/2{
			for range x/2-len(p)/2{
				lines[i] += " "
			}
			lines[i] += cWhite + p + hdr
			for range int(math.Ceil(float64(x)/2))-int(math.Ceil(float64(len(p))/2)) - 1{
				lines[i] += " "
			}
		}else{
			for range x-1{
				lines[i] += " "
			}
		}
		lines[i] += hdr + "▐" + reset + "" 
	}
	lines = append(lines, hdr)
	for range x-1{
		lines[len(lines)-1] += "▄"
	}
	lines[len(lines)-1] += "▟" + reset + ""
	return shape{x: x, y: y, lines: lines}
}

func appendShapesVertically(r1, r2 shape) shape{
	if r1.x == 0 {
		return r2
	}

	return shape{
		x: r1.x,
		y: r1.y + r2.y,
		lines: append(r2.lines, r1.lines...),
	}
}

func appendShapesHorizontally(r1, r2 shape) shape{
	if r1.x == 0 {
		return r2
	}
	r := shape{r2.x + r1.x, r2.y, r2.lines}

	for i, l := range r1.lines{
		r.lines[i] += l
	}
	return r
}

func buildStocks(x, y int, s []stockmodel, totalMarketCap float64, hdr string, hideHdrs bool) shape{
	if len(s) == 0 || x == 0 || y == 0{
		return shape{}
	}

	if y == 1 || hideHdrs{
		hdr = ""
	}
	if hdr != "" {
		y--
	}

	a := x*y
	w := s[0].marketCap/totalMarketCap
	size := w*float64(a)

	if w > 0.33 {
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			return addHeader(
				hdr,
				appendShapesHorizontally(
					buildStocks(x-rx, y, s[1:], totalMarketCap-s[0].marketCap, "", hideHdrs),
					rect(rx, y, s[0].color, s[0].ticker, s[0].performance),
				),
			)
		}
		ry := int(math.Ceil(size/float64(x)))
		return addHeader(
			hdr,
			appendShapesVertically(
				buildStocks(x, y-ry, s[1:], totalMarketCap-s[0].marketCap, "", hideHdrs),
				rect(x, ry, s[0].color, s[0].ticker, s[0].performance),
			),
		)
	}else{
		w = (s[1].marketCap + s[0].marketCap)/totalMarketCap
		w1 := s[0].marketCap / (s[0].marketCap + s[1].marketCap)
		size = w * float64(a)
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			y1 := int(math.Ceil(float64(y)*w1))
			return addHeader(
				hdr,
				appendShapesHorizontally(
					buildStocks(x-rx, y, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap, "", hideHdrs),
					appendShapesVertically(
						rect(rx, y-y1, s[1].color, s[1].ticker, s[0].performance),
						rect(rx, y1, s[0].color, s[0].ticker, s[0].performance),
					),
				),
			)
		}else {
			ry := int(math.Ceil(size/float64(x)))
			x1 := int(math.Ceil(float64(x)*w1))
			return addHeader(
				hdr,
				appendShapesVertically(
					buildStocks(x, y-ry, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap, "", hideHdrs),
					appendShapesHorizontally(
						rect(x-x1, ry, s[1].color, s[1].ticker, s[0].performance),
						rect(x1, ry, s[0].color, s[0].ticker, s[0].performance),
					),
				),
			)
		}
	}
}

func addHeader(hdr string, s shape) shape {
	if hdr == "" {
		return s
	}

	if len(hdr) <= s.x {
		hdr += strings.Repeat(" ", s.x - len(hdr))
	}else {
		hdr = hdr[:s.x]
	}
	return shape{
		x: s.x,
		y: s.y + 1,
		lines: append([]string{hdr}, s.lines...),
	}
}

func buildIndustries(x,y int, i []industrymodel, totalMarketCap float64, hdr string, hideHdrs bool) shape {
	if len(i) == 0 || x == 0 || y == 0{
		return shape{}
	}

	if y == 1 {
		hdr = ""
	}
	if hdr != "" {
		y--
	}

	a := x*y
	w := i[0].marketCap/totalMarketCap
	size := w*float64(a)

	out := shape{}
	if w > 0.33 {
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			out = appendShapesHorizontally(
				buildIndustries(x-rx, y, i[1:], totalMarketCap-i[0].marketCap, "", hideHdrs),
				buildStocks(rx, y, i[0].stocks, i[0].marketCap, i[0].name, hideHdrs),
			)
		}else{
			ry := int(math.Ceil(size/float64(x)))
			out = appendShapesVertically(
				buildIndustries(x, y-ry, i[1:], totalMarketCap-i[0].marketCap, "", hideHdrs),
				buildStocks(x, ry, i[0].stocks, i[0].marketCap, i[0].name, hideHdrs),
			)
		}
	}else{
		w = (i[1].marketCap + i[0].marketCap)/totalMarketCap
		w1 := i[0].marketCap / (i[0].marketCap + i[1].marketCap)
		size = w * float64(a)
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			y1 := int(math.Ceil(float64(y)*w1))
			out = appendShapesHorizontally(
				buildIndustries(x-rx, y, i[2:], totalMarketCap-i[0].marketCap-i[1].marketCap, "", hideHdrs),
				appendShapesVertically(
					buildStocks(rx, y-y1, i[1].stocks, i[1].marketCap, i[1].name, hideHdrs),
					buildStocks(rx, y1, i[0].stocks, i[0].marketCap, i[0].name, hideHdrs),
				),
			)
		}else {
			ry := int(math.Ceil(size/float64(x)))
			x1 := int(math.Ceil(float64(x)*w1))
			out = appendShapesVertically(
				buildIndustries(x, y-ry, i[2:], totalMarketCap-i[0].marketCap-i[1].marketCap, "", hideHdrs),
				appendShapesHorizontally(
					buildStocks(x-x1, ry, i[1].stocks, i[1].marketCap, i[1].name, hideHdrs),
					buildStocks(x1, ry, i[0].stocks, i[0].marketCap, i[0].name, hideHdrs),
				),
			)
		}
	}
	return addHeader(hdr, out)
}

func buildSectors(x,y int, s []sectormodel, totalMarketCap float64, hdr string) shape {
	if len(s) == 0 || x == 0 || y == 0{
		return shape{}
	}

	if hdr != "" {
		y--
	}

	a := x*y
	w := s[0].marketCap/totalMarketCap
	size := w*float64(a)

	out := shape{}
	if w > 0.33 {
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			out = appendShapesHorizontally(
				buildSectors(x-rx, y, s[1:], totalMarketCap-s[0].marketCap, ""),
				buildIndustries(rx, y, s[0].industries, s[0].marketCap, s[0].name, true),
			)
		}else{
			ry := int(math.Ceil(size/float64(x)))
			out = appendShapesVertically(
				buildSectors(x, y-ry, s[1:], totalMarketCap-s[0].marketCap, ""),
				buildIndustries(x, ry, s[0].industries, s[0].marketCap, s[0].name, true),
			)
		}
	}else{
		w = (s[1].marketCap + s[0].marketCap)/totalMarketCap
		w1 := s[0].marketCap / (s[0].marketCap + s[1].marketCap)
		size = w * float64(a)
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			y1 := int(math.Ceil(float64(y)*w1))
			out = appendShapesHorizontally(
				buildSectors(x-rx, y, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap, ""),
				appendShapesVertically(
					buildIndustries(rx, y-y1, s[1].industries, s[1].marketCap, s[1].name, true),
					buildIndustries(rx, y1, s[0].industries, s[0].marketCap, s[0].name, true),
				),
			)
		}else {
			ry := int(math.Ceil(size/float64(x)))
			x1 := int(math.Ceil(float64(x)*w1))
			out = appendShapesVertically(
				buildSectors(x, y-ry, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap, ""),
				appendShapesHorizontally(
					buildIndustries(x-x1, ry, s[1].industries, s[1].marketCap, s[1].name, true),
					buildIndustries(x1, ry, s[0].industries, s[0].marketCap, s[0].name, true),
				),
			)
		}
	}
	return addHeader(hdr, out)
}

func main() {
	f, err := os.Open("spy.json")

	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	b, _ := io.ReadAll(f)
	var stocks Stocks

	json.Unmarshal(b, &stocks)

	stocks.totalCap = 0
	for _, s := range stocks.Stocks{
		stocks.totalCap += s.MarketCap
	}

	p := tea.NewProgram((initialModel(stocks)))
	if _, err := p.Run(); err != nil {
		fmt.Printf("error")
		os.Exit(1)
	}
}
