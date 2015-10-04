package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
	"os"
	"strconv"	
)

type RequestObject struct {
	StockKey          string
	Budget            float32
}

type ResponseObject struct {
	TradeID			 int
	Stocks           []string
	UnvestedAmount   float32
	
}

type PortfolioResponseObject struct {
	Stocks    []string
	CurrentMarketValue float32
	UnvestedAmount float32
}


func main() {
  if len(os.Args) == 3 {
        buyStocks()
    }else if len(os.Args) == 2{
        checkPortfolio()
  }else {
        fmt.Println("Zzz:", os.Args[0], "127.0.0.1:1234")
        log.Fatal(1)
  }	
}

func buyStocks() {

	var requestObject RequestObject
	
	requestObject.StockKey = os.Args[1]
    budget64, _  := strconv.ParseFloat(os.Args[2], 32)
    requestObject.Budget = float32(budget64)


	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }
	
	respObj := new(ResponseObject)
	
	err = client.Call("StockObject.ParseRequestObject", requestObject, &respObj)

	if err != nil {
		log.Fatal("Error thrown!", err)
	}
	
	fmt.Print("Result:")
	fmt.Println(respObj.Stocks)
	fmt.Print("TradeID:")
	fmt.Println(respObj.TradeID)
	fmt.Print("UnvestedAmount:")
	fmt.Println(respObj.UnvestedAmount)
	
}	


func checkPortfolio() {

	var TradeID int

    tradeID,_ := strconv.ParseInt(os.Args[1], 10, 32)
    TradeID = int(tradeID)
	
    client, err := jsonrpc.Dial("tcp", "localhost:1234")
    if err != nil {
        log.Fatal("dialing:", err)
    }
    
    var pfResponseObj PortfolioResponseObject
    
        err = client.Call("StockObject.CheckPortfolio", TradeID, &pfResponseObj)
        if err != nil {
            log.Fatal("Wrong Input", err)
        }

        fmt.Print("Stocks:")
        fmt.Println(pfResponseObj.Stocks)        
        fmt.Print("Current Market Value:")
		fmt.Println(pfResponseObj.CurrentMarketValue)
        fmt.Print("Unvested Amount:")
		fmt.Println(pfResponseObj.UnvestedAmount)
}	

