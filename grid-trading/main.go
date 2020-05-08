package main

import (
	. "github.com/coinrust/crex"
	"github.com/coinrust/crex/serve"
	"log"
	"time"
)

type Level struct {
	Price      float64
	HoldPrice  float64
	HoldAmount float64
	CoverPrice float64
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

func GridShift(grid *[]Level) *Level {
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

type GridStrategy struct {
	StrategyBase

	Grid []Level

	StopLoss float64
	StopWin  float64

	Symbol    string  `opt:"symbol,BTUSDT"`
	Direction float64 `opt:"direction,1"` // 网格方向 up 1, down -1

	GridNum         int     `opt:"grid_num,10"`         // 网格节点数量 10
	GridPointAmount float64 `opt:"grid_point_amount,1"` // 网格节点下单量 1
	GridPointDis    float64 `opt:"grid_point_dis,20"`   // 网格节点间距 20
	GridCovDis      float64 `opt:"grid_cov_dis,50"`     // 网格节点平仓价差 50
}

func (s *GridStrategy) OnInit() error {
	log.Printf("Symbol: %v", s.Symbol)
	log.Printf("Direction: %v", s.Direction)
	log.Printf("GridNum: %v", s.GridNum)
	log.Printf("GridPointAmount: %v", s.GridPointAmount)
	log.Printf("GridPointDis: %v", s.GridPointDis)
	log.Printf("GridCovDis: %v", s.GridCovDis)
	return nil
}

func (s *GridStrategy) OnTick() error {
	ob, err := s.Exchange.GetOrderBook(s.Symbol, 1)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	s.UpdateGrid(ob)
	return nil
}

func (s *GridStrategy) UpdateGrid(ob *OrderBook) {
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

		var order *Order
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
		var order *Order
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
		var order *Order
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
		GridShift(&s.Grid)
		s.StopLoss++
	}
}

func (s *GridStrategy) Run() error {
	// run loop
	for {
		s.OnTick()
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (s *GridStrategy) OnExit() error {
	return nil
}

func main() {
	s := &GridStrategy{}
	err := serve.Serve(s)
	if err != nil {
		log.Printf("%v", err)
	}
}
