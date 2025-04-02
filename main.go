package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"github.com/joho/godotenv"
	"net/http"
	// tea "github.com/charmbracelet/bubbletea"
)

type Stocks struct {
	Stocks []StockInfo `json:"stocks"`
}

type StockInfo struct {
	Ticker string `json:"ticker"`
	Name string `json:"name"`
	Sector string `json:"sector"`
	Industry string `json:"industry"`
	Weight float64
}

func main(){
	f, err := os.Open("spy.json")

	if err != nil{
		fmt.Println(err)
	}

	defer f.Close()

	b, _ := io.ReadAll(f)
	var stocks Stocks

	json.Unmarshal(b, &stocks)

	// for i := 0; i < len(stocks.Stocks); i++ {
	// 	fmt.Println("Ticker = " + stocks.Stocks[i].Ticker)
	// 	fmt.Println("Name = " + stocks.Stocks[i].Name)
	// 	fmt.Println("Sector = " + stocks.Stocks[i].Sector)
	// 	fmt.Println("Industry = " + stocks.Stocks[i].Industry)
		
	// }

	envFile, _ := godotenv.Read(".env")

	API_KEY := envFile["API_KEY"]


	stock, err := http.Get("https://www.alphavantage.co/query?function=ETF_PROFILE&symbol=SPDR&apikey="+API_KEY)

	if err != nil{
		panic(err)
	}

	fmt.Println(stock.)
}