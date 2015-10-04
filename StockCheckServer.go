package main

import (
	"fmt"
	"strings"
	"strconv"
	"net"
	"net/rpc"
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
	"net/rpc/jsonrpc"
    "os"
	"math"
	"math/rand"
	"errors"
	

)


type StockObject struct {
	StockPF map[int](*PortfolioObject)
}

type PortfolioObject struct {
	Stocks           map[string](*ShareObject)
	UnvestedAmount float32
}

type ShareObject struct {
	PurchasedPrice float32
	SharesCount    int
}

type RequestObject struct {
	StockKey string
	Budget   float32
}

type ResponseObject struct {
	TradeID    int
	Stocks     []string
	UnvestedAmount float32
}

type PortfolioResponseObject struct {
	Stocks    []string
	CurrentMarketValue float32
	UnvestedAmount float32
	
}



var tradeID int

type JSONObj struct {
    List struct {
        Resources []struct {
            Resource struct {
                Fields struct {
                    Name    string `json:"name"`
                    Price   string `json:"price"`
                    Symbol  string `json:"symbol"`
                    Ts      string `json:"ts"`
                    Type    string `json:"type"`
                    UTCTime string `json:"utctime"`
                    Volume  string `json:"volume"`
                } `json:"fields"`
            } `json:"resource"`
        } `json:"resources"`
    } `json:"list"`
}


func (so *StockObject) ParseRequestObject(args *RequestObject, rsp *ResponseObject) error {
	
	tradeID++
	rsp.TradeID = tradeID
	
	
	if so.StockPF == nil {

		so.StockPF = make(map[int](*PortfolioObject))

		so.StockPF[tradeID] = new(PortfolioObject)
		so.StockPF[tradeID].Stocks = make(map[string]*ShareObject)

	}
	
	stockAndPercentages := strings.Split(args.StockKey, ",")
	
	fmt.Println("The Budget is", args.Budget)
	
	budget := float32(args.Budget)
	
	var totalSpent float32
	
	for _, stocks := range stockAndPercentages {

		splitString := strings.Split(stocks, ":")
		stockSymbol := splitString[0]
		stockPercentage := splitString[1]
		stockPercent := strings.TrimSuffix(stockPercentage, "%")
		fPercent64, _ := strconv.ParseFloat(stockPercent, 32)
		fPercent := float32(fPercent64 / 100.00)
		
		fmt.Println("The Stock Symbol is",stockSymbol)
		
		fmt.Println("It's Percentage is", fPercent)
		
		financeAPIPrice := callYahooAPI(stockSymbol)
		
		sharesCount := int(math.Floor(float64(budget * fPercent / financeAPIPrice)))
		sharesCountFloat := float32(sharesCount)
		totalSpent += sharesCountFloat * financeAPIPrice

		endResult := stockSymbol + ":" + strconv.Itoa(sharesCount) + ":$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)

		rsp.Stocks = append(rsp.Stocks, endResult)
		
		
		if _, ok := so.StockPF[tradeID]; !ok {

			pfObj := new(PortfolioObject)
			pfObj.Stocks = make(map[string]*ShareObject)
			so.StockPF[tradeID] = pfObj
		}
		if _, ok := so.StockPF[tradeID].Stocks[stockSymbol]; !ok {

			shareObj := new(ShareObject)
			shareObj.PurchasedPrice = financeAPIPrice
			shareObj.SharesCount = sharesCount
			so.StockPF[tradeID].Stocks[stockSymbol] = shareObj
		} else {

			total := float32(sharesCountFloat*financeAPIPrice) + float32(so.StockPF[tradeID].Stocks[stockSymbol].SharesCount)*so.StockPF[tradeID].Stocks[stockSymbol].PurchasedPrice
			so.StockPF[tradeID].Stocks[stockSymbol].PurchasedPrice = total / float32(sharesCount+so.StockPF[tradeID].Stocks[stockSymbol].SharesCount)
			so.StockPF[tradeID].Stocks[stockSymbol].SharesCount += sharesCount
		}
		
		
	}
	
	unvestedAmount := budget - totalSpent
	rsp.UnvestedAmount = unvestedAmount
	so.StockPF[tradeID].UnvestedAmount += unvestedAmount

	
	return nil
}



func (so* StockObject) CheckPortfolio(tradeID int ,  rsp *PortfolioResponseObject) error {
						

				if objValues, ok := so.StockPF[tradeID]; ok {

					var currentMarketValue float32
					for stockSymbol, so := range objValues.Stocks {
						
						financeAPIPrice := callYahooAPI(stockSymbol)

						var result string
						if so.PurchasedPrice < financeAPIPrice {
							result = "+$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
						} else if so.PurchasedPrice > financeAPIPrice {
							result = "-$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
						} else {
							result = "$" + strconv.FormatFloat(float64(financeAPIPrice), 'f', 2, 32)
						}
						stock := stockSymbol + ":" + strconv.Itoa(so.SharesCount) + ":" + result

						rsp.Stocks = append(rsp.Stocks, stock)

						currentMarketValue += float32(so.SharesCount) * financeAPIPrice
					}
					fmt.Print("Unvested amount zz is", objValues.UnvestedAmount)
					rsp.UnvestedAmount = objValues.UnvestedAmount
					rsp.CurrentMarketValue = currentMarketValue
				}else {
					return errors.New("Trade ID doesnt exist")
				}

				return nil
	}
	
	
	
func main() {
	
	tradeID = rand.Intn(100) + 1
	stocksystem := new(StockObject)
	rpc.Register(stocksystem)
    tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
    checkError(err)

    listener, err := net.ListenTCP("tcp", tcpAddr)
    checkError(err)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        jsonrpc.ServeConn(conn)
    }

}


func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}

func callYahooAPI(stockSymbol string) float32 {
	
	
	url := fmt.Sprintf("http://finance.yahoo.com/webservice/v1/symbols/%s/quote?format=json",stockSymbol)
	
	urlRes,err := http.Get(url)
	
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(urlRes.Body)
	urlRes.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	var jsonObj JSONObj

    err = json.Unmarshal(body, &jsonObj)
	
	if err != nil{
            panic(err)
        }
    
	
	fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Name)
    fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Symbol)
    fmt.Println(jsonObj.List.Resources[0].Resource.Fields.Price)
	
	floatFinalPrice, err := strconv.ParseFloat(jsonObj.List.Resources[0].Resource.Fields.Price, 32)

	return float32(floatFinalPrice)
}





