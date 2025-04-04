package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

var reset = "\033[0m"
var bold = "\033[1m"
var underline = "\033[4m"
var strike = "\033[9m"
var italic = "\033[3m"

var cRed = "\033[31m"
var cGreen = "\033[32m"
var cYellow = "\033[33m"
var cBlue = "\033[34m"
var cPurple = "\033[35m"
var cCyan = "\033[36m"
var cWhite = "\033[37m"

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
	industries []industrymodel
}

type industrymodel struct {
	name string
	stocks []stockmodel
}

type stockmodel struct {
	ticker string
	performance float64
	weight float64
}

func initialModel(stocks Stocks) model {
	m := model{sectors: []sectormodel{}}
	
	for i, s := range stocks.Stocks {
		newStock := stockmodel{
			ticker: s.Ticker,
			weight: float64(s.MarketCap)/float64(stocks.totalCap), 
			performance: (s.CurrentPrice - s.OpenPrice) / s.OpenPrice,
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
	// s := "What should we buy at the market \n\n"

	// for i, choice := range m.choices{
	// 	cursor := " "
	// 	if m.cursor == i {
	// 		cursor = ">"
	// 	}

	// 	checked := " "
	// 	if _, ok := m.selected[i]; ok{
	// 		checked = "x"
	// 	}

	// 	s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	// }

	// s := cRed + "████████" + reset + "████████"
	// s += cGreen + "████████" + bold + "████████"
	// s += reset
	str := ""
	for _, s := range m.sectors[2].industries[3].stocks{
		c := cRed
		if s.performance > 0.02 {
			c = cGreen
		} else if s.performance > 0.0075 {
			c = cGreen + bold
		} else if s.performance > -0.0075 {
			c = cWhite
		} else if s.per + bold
		}
		str += reset + s.ticker + "\n" + c + strconv.FormatFloat(s.performance, 'f', -1, 64) + "\n" + "████████\n" + strconv.FormatFloat(s.weight, 'f', -1, 64) + "\n\n"
	}

	return str
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
