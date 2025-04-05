package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

var reset = "\033[0m"
var bold = "\033[1m"
var underline = "\033[4m"
var strike = "\033[9m"
var italic = "\033[3m"

var cRed = "\033[31m"
var cLightRed = "\033[91m"
var cGreen = "\033[32m"
var cLightGreen = "\033[92m"
var cGrey = "\033[90m"
var cBlack = "\033[30m"

var cRedBg = "\033[41m"
var cLightRedBg = "\033[101m"
var cGreenBg = "\033[42m"
var cLightGreenBg = "\033[102m"
var cGreyBg = "\033[100m"
var cBlackBg = "\033[30m"

var cYellow = "\033[33m"
var cBlue = "\033[34m"
var cPurple = "\033[35m"
var cCyan = "\033[36m"
var cWhite = "\033[97m"
var cLightGrey = "\033[37m"

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
}

type sectormodel struct {
	name string
	weight float64
	marketCap float64
	industries []industrymodel
}

type industrymodel struct {
	name string
	weight float64
	marketCap float64
	stocks []stockmodel
}

type stockmodel struct {
	ticker string
	performance float64
	weight float64
	marketCap float64
	color string
}

type shape struct {
	x int
	y int
	lines []string
}

func initialModel(stocks Stocks) model {
	m := model{sectors: []sectormodel{}}
	
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

			for k, stock := range ind.stocks {
				ind.stocks[k].weight = stock.marketCap/indCap
			}
			sec.industries[j].marketCap = indCap
			secCap += indCap
		}
		for j, ind := range sec.industries {
			sec.industries[j].weight = ind.marketCap/secCap
		}

		m.sectors[i].marketCap = secCap
	}
	for i, sec := range m.sectors {
		m.sectors[i].weight = sec.marketCap/float64(stocks.totalCap)
	}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type){
	case tea.KeyMsg:
		switch msg.String(){
		case "ctrl+c", "q":
			return m, tea.Quit

	// 	case "up", "k":
	// 		if m.cursor > 0{
	// 			m.cursor--
	// 		}
	// 	case "down", "j":
	// 		if m.cursor < len(m.choices)-1{
	// 			m.cursor++
	// 		}
	// 	case "enter", " ":
	// 		_, ok := m.selected[m.cursor]
	// 		if ok {
	// 			delete(m.selected, m.cursor)
	// 		} else{
	// 			m.selected[m.cursor] = struct{}{}
	// 		}
		}
	}
	return m, nil
}

func (m model) View() string {
	str := ""
	
	s := m.sectors[9].industries[6].stocks
	slices.SortFunc(s, func(i,j stockmodel) int{
		return -cmp.Compare(i.weight, j.weight)
	})

	stocks := buildStocks(50, 30, s, m.sectors[9].industries[6].marketCap)

	for _, l := range stocks.lines{
		str += l + "\n"
	}
	str += m.sectors[9].industries[6].name

	// // slices.Reverse(stocks)
	// box := fillStocks(15,10,stocks)
	// for _, l := range box.lines {
	// 	str += l + "\n"
	// }
	// str +=fmt.Sprint(box.x,box.y) + "\n"

	// for _, s := range stocks{
	// 	str += fmt.Sprint(s.x) + " " + fmt.Sprint(s.y) + "\n"
	// }
	// str = ""
	// for _, sh := range buildStocks(15, 10, s, m.sectors[2].industries[1].marketCap){
	// 	for _, l := range sh.lines {
	// 		str += l + "\n"
	// 	}
	// 	str+="\n"
	// }
	// str = ""
	// for _, s := range m.sectors{
	// 	str += s.name + "\n"
	// 	for _, i := range s.industries{
	// 		str += i.name + "\n"
	// 		for _, st := range i.stocks{
	// 			str += st.color + "1"
	// 		}
	// 	}
	// 	str += "\n"
	// }
	return str
}


// create a rectangle of width x - 0.5 and height y - 0.5
// this rectangle already has the bottom and right gaps on it hence the - 0.5
// color the rectangle cbg
// uses background colors and makes the foreground black to match the terminal (dont need to do this anymore but already wrote the code)
func rect(x, y int, cbg, t string) shape{
	hdr := cbg + cBlack
	if x == 0 || y == 0 {
		return shape{0,0,[]string{}}
	}

	lines := make([]string, y-1)
	for i := range lines {
		lines[i] += hdr
		if x == 4 && len(t) == 3 && i == y/2-1{
			lines[i] += cWhite + t + hdr
		}else if x > len(t) && i == y/2-1{
			for range x/2-len(t)/2{
				lines[i] += " "
			}
			lines[i] += cWhite + t + hdr
			for range int(math.Ceil(float64(x)/2))-int(math.Ceil(float64(len(t))/2)) - 1{
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
	// if r1.x == r2.x{
	// 	if r1.y > r2.y{
	// 		return shape{
	// 			x: r1.x,
	// 			y: r1.y + r2.y,
	// 			lines: append(r1.lines, r2.lines...),
	// 		}
	// 	}else {
	// 		return shape{
	// 			x: r2.x,
	// 			y: r1.y + r2.y,
	// 			lines: append(r2.lines, r1.lines...),
	// 		}
	// 	}
	// }else if r1.x > r2.x {
	// 	return shape{
	// 		x: r1.x,
	// 		y: r1.y + r2.y,
	// 		lines: append(r1.lines, r2.lines...),
	// 	}
	// }else {
	// 	return shape{
	// 		x: r2.x,
	// 		y: r1.y + r2.y,
	// 		lines: append(r2.lines, r1.lines...),
	// 	}
	// }
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
// func appendShapes(r1, r2 shape) shape{
// 	fmt.Print("{", r1.x, r1.y, r2.x, r2.y, "}")
// 	if r1.x == r2.x{
// 		fmt.Println("vert")
// 		return appendShapesVertically(r1, r2)
// 	}else if r1.y == r2.y {
// 		fmt.Println("hor")
// 		return appendShapesHorizontally(r1, r2)
// 	}
// 	return r2
// }
// func fillStocks(x, y int, shapes []shape) shape{
// 	tot := shapes[0]
// 	for i,s := range shapes[1:]{
// 		fmt.Print(i, tot.x, tot.y, s.x, s.y)
// 		tot = appendShapes(tot, s)
// 	}
// 	fmt.Println()
// 	return tot
// }

func buildStocks(x, y int, s []stockmodel, totalMarketCap float64) shape{
	if len(s) == 0 {
		return shape{}
	}
	a := x*y
	w := s[0].marketCap/totalMarketCap
	size := w*float64(a)

	if w > 0.33 {
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			return appendShapesHorizontally(
				buildStocks(x-rx, y, s[1:], totalMarketCap-s[0].marketCap),
				rect(rx, y, s[0].color, s[0].ticker),
			)
		}
		ry := int(math.Ceil(size/float64(x)))
		return appendShapesVertically(
			buildStocks(x, y-ry, s[1:], totalMarketCap-s[0].marketCap),
			rect(x, ry, s[0].color, s[0].ticker),
		)
	}else{
		w = (s[1].marketCap + s[0].marketCap)/totalMarketCap
		w1 := s[0].marketCap / (s[0].marketCap + s[1].marketCap)
		size = w * float64(a)
		if x >= y {
			rx := int(math.Ceil(size/float64(y)))
			y1 := int(math.Ceil(float64(y)*w1))
			return appendShapesHorizontally(
				buildStocks(x-rx, y, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap),
				appendShapesVertically(
					rect(rx, y-y1, s[1].color, s[1].ticker),
					rect(rx, y1, s[0].color, s[0].ticker),
				),
			)
		}else {
			ry := int(math.Ceil(size/float64(x)))
			x1 := int(math.Ceil(float64(x)*w1))
			return appendShapesVertically(
				buildStocks(x, y-ry, s[2:], totalMarketCap-s[0].marketCap-s[1].marketCap),
				appendShapesHorizontally(
					rect(x-x1, ry, s[1].color, s[1].ticker),
					rect(x1, ry, s[0].color, s[0].ticker),
				),
			)
		}
	}
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
