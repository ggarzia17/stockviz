import requests
from bs4 import BeautifulSoup
import json


WIKI = 'https://en.wikipedia.org/wiki/List_of_S&P_500_companies'

class StockInfo():
    ticker: str
    name: str
    sector: str
    industry: str
    def __init__(self):
        self.ticker = ""
        self.name = ""
        self.sector = ""
        self.industry = ""
    def toString(self):
        return f"StockInfo{{ticker = {self.ticker}, name = {self.name}, sector = {self.sector}, industry = {self.industry}}}"

def get_spy_tickers() -> StockInfo:
    soup = BeautifulSoup(requests.get(WIKI).text, 'html.parser')

    # print(soup.findAll("table", {'class': 'wikitable sortable sticky-header'}))
    symbols: list[StockInfo] = []
    iTicker = -1
    iName = -1
    iSector = -1
    iIndustry = -11

    for table in soup.findAll("table", {'class': 'wikitable sortable sticky-header'}):
        header = table.findAll('th')
        for i in range(len(header)):
            s = header[i].text.lower()
            print(s)
            if 'symbol' in s:
                iTicker = i
            if 'security' in s:
                iName = i
            if 'gics sector' in s:
                iSector = i
            if 'gics sub-industry' in s:
                iIndustry = i

        if iTicker != -1:
            for row in table.findAll('tr'):
                symbols.append(StockInfo())
                fields = row.findAll('td')
                if fields and fields[iTicker]:
                    symbol = fields[iTicker].text.strip()
                    if ':' in symbol:
                        symbol = symbol.split(':')[1].strip()
                    symbols[-1].ticker = symbol
                if fields and fields[iName]:
                    symbol = fields[iName].text.strip()
                    if ':' in symbol:
                        symbol = symbol.split(':')[1].strip()
                    symbols[-1].name = symbol
                if fields and fields[iSector]:
                    symbol = fields[iSector].text.strip()
                    if ':' in symbol:
                        symbol = symbol.split(':')[1].strip()
                    symbols[-1].sector = symbol
                if fields and fields[iIndustry]:
                    symbol = fields[iIndustry].text.strip()
                    if ':' in symbol:
                        symbol = symbol.split(':')[1].strip()
                    symbols[-1].industry = symbol
                
                
            break
    spyTickers = []
    for s in symbols:
        if s is not None and s.ticker and s.industry and s.name and s.sector:
            spyTickers += [s]
    return spyTickers



if __name__ == '__main__':
    l = {"stocks": []}
    for s in get_spy_tickers():
        l["stocks"] += [{
            "ticker": s.ticker,
            "name": s.name,
            "sector": s.sector,
            "industry": s.industry
        }]
    with open("../spy.json", "w") as f:
        json.dump(l, f)
