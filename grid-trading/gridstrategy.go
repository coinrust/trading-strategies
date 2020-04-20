package main

import (
	. "github.com/coinrust/crex"
	"github.com/coinrust/crex/exchanges"
	"github.com/spf13/viper"
	"log"
	"time"
)

type Level struct {
	Price      float64
	HoldPrice  float64
	HoldAmount float64
	CoverPrice float64
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
}

func GridPop(grid *[]Level) *Level {
	length := len(*grid)
	if length == 0 {
		return nil
	}
	item := (*grid)[length-1]
	*grid = (*grid)[:length-1]
	return &item
}

func Shift(grid *[]Level) *Level {
	length := len(*grid)
	if length == 0 {
		return nil
	}
	item := (*grid)[0]
	if length > 1 {
		*grid = (*grid)[1:length]
	} else {
		*grid = []Level{}
	}
	return &item
}

type BasicStrategy struct {
	StrategyBase

	Grid []Level

	StopLoss float64
	StopWin  float64

	Symbol    string
	Direction float64 // 网格方向 up 1, down -1

	GridNum         int     // 网格节点数量 10
	GridPointAmount float64 // 网格节点下单量 1
	GridPointDis    float64 // 网格节点间距 20
	GridCovDis      float64 // 网格节点平仓价差 50
}

func (s *BasicStrategy) OnInit() {
	s.Symbol = viper.GetString("symbol")
	s.Direction = viper.GetFloat64("direction")
	s.GridNum = viper.GetInt("grid_num")
	s.GridPointAmount = viper.GetFloat64("grid_point_amount")
	s.GridPointDis = viper.GetFloat64("grid_point_dis")
	s.GridCovDis = viper.GetFloat64("grid_cov_dis")
}

func (s *BasicStrategy) OnTick() {
	ob, err := s.Exchange.GetOrderBook(s.Symbol, 1)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	s.UpdateGrid(&ob)
}

func (s *BasicStrategy) UpdateGrid(ob *OrderBook) {
	nowAskPrice, nowBidPrice := ob.AskPrice(), ob.BidPrice()
	if len(s.Grid) == 0 ||
		(s.Direction == 1 && nowBidPrice-s.Grid[len(s.Grid)-1].Price > s.GridCovDis) ||
		(s.Direction == -1 && s.Grid[len(s.Grid)-1].Price-nowAskPrice > s.GridCovDis) {

		nowPrice := nowAskPrice
		if s.Direction == 1 {
			nowPrice = nowBidPrice
		}
		price := nowPrice
		coverPrice := nowPrice - s.Direction*s.GridCovDis
		if len(s.Grid) > 0 {
			price = s.Grid[len(s.Grid)-1].Price + s.GridPointDis*s.Direction
			coverPrice = s.Grid[len(s.Grid)-1].Price + s.GridPointDis*s.Direction*s.GridCovDis
		}

		s.Grid = append(s.Grid, Level{
			Price:      price,
			HoldPrice:  0,
			HoldAmount: 0,
			CoverPrice: coverPrice,
		})

		var order Order
		var err error
		if s.Direction == 1 {
			order, err = s.Exchange.OpenShort(s.Symbol, OrderTypeMarket, 0, s.GridPointAmount)
		} else {
			order, err = s.Exchange.OpenLong(s.Symbol, OrderTypeMarket, 0, s.GridPointAmount)
		}
		if err != nil {
			log.Printf("%v", err)
			return
		}

		log.Printf("委托成交 ID=%v 成交价=%v 成交数量=%v Direction=%v",
			order.ID, order.AvgPrice, order.FilledAmount, s.Direction)

		s.Grid[len(s.Grid)-1].HoldPrice = order.AvgPrice
		s.Grid[len(s.Grid)-1].HoldAmount = order.FilledAmount
	}
	if len(s.Grid) > 0 &&
		((s.Direction == 1 && nowAskPrice < s.Grid[len(s.Grid)-1].CoverPrice) ||
			(s.Direction == -1 && nowBidPrice > s.Grid[len(s.Grid)-1].CoverPrice)) {
		var order Order
		var err error
		size := s.Grid[len(s.Grid)-1].HoldAmount
		if s.Direction == 1 {
			order, err = s.Exchange.OpenLong(s.Symbol, OrderTypeMarket, 0, size)
		} else {
			order, err = s.Exchange.OpenShort(s.Symbol, OrderTypeMarket, 0, size)
		}
		if err != nil {
			log.Printf("%v", err)
			return
		}
		log.Printf("order=%#v", order)
		GridPop(&s.Grid)
		s.StopWin++
	} else if len(s.Grid) > s.GridNum {
		var order Order
		var err error
		size := s.Grid[0].HoldAmount
		if s.Direction == 1 {
			order, err = s.Exchange.OpenLong(s.Symbol, OrderTypeMarket, 0, size)
		} else {
			order, err = s.Exchange.OpenShort(s.Symbol, OrderTypeMarket, 0, size)
		}
		if err != nil {
			log.Printf("%v", err)
			return
		}
		log.Printf("order=%#v", order)
		Shift(&s.Grid)
		s.StopLoss++
	}
}

func (s *BasicStrategy) OnDeinit() {

}

func main() {
	exchangeID := viper.GetString("exchange_id")
	accessKey := viper.GetString("access_key")
	secretKey := viper.GetString("secret_key")
	testnet := viper.GetBool("testnet")

	broker := exchanges.NewExchange(exchangeID,
		ApiAccessKeyOption(accessKey),
		ApiSecretKeyOption(secretKey),
		ApiTestnetOption(testnet))

	s := &BasicStrategy{}
	s.Setup(TradeModeLiveTrading, broker)

	// run loop
	for {
		s.OnTick()
		time.Sleep(500 * time.Millisecond)
	}
}
