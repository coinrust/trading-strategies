package main

import (
	"fmt"
	. "github.com/coinrust/crex"
	"github.com/coinrust/crex/serve"
	"log"
	"time"
)

type Hold struct {
	Price  float64
	Amount float64
}

// 赌徒策略(做空)
type GamblerStrategy struct {
	StrategyBase

	StopWin     float64 `opt:"stop_win,500"`   // 止盈
	StopLoss    float64 `opt:"stop_loss,500"`  // 止损
	FirstAmount float64 `opt:"first_amount,1"` // 单次下单量
	MaxGear     int     `opt:"max_gear,8"`     // 加倍赌的次数

	Currency string `opt:"currency,BTC"`  // 货币
	Symbol   string `opt:"symbol,BTCUSD"` // 标

	hold Hold
	gear int
}

func (s *GamblerStrategy) OnInit() error {
	log.Printf("StopWin: %v", s.StopWin)
	log.Printf("StopLoss: %v", s.StopLoss)
	log.Printf("FirstAmount: %v", s.FirstAmount)
	log.Printf("MaxGear: %v", s.MaxGear)
	log.Printf("Currency: %v", s.Currency)
	log.Printf("Symbol: %v", s.Symbol)

	balance, err := s.Exchange.GetBalance(s.Currency)
	if err != nil {
		log.Printf("%v", err)
		return err
	}
	log.Printf("初始资产 Equity: %v Available: %v",
		balance.Equity, balance.Available)
	return nil
}

func (s *GamblerStrategy) OnTick() (err error) {
	var ob OrderBook
	ob, err = s.Exchange.GetOrderBook(s.Symbol, 1)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if s.hold.Amount == 0 {
		var order *Order
		order, err = s.Sell(s.FirstAmount)
		s.hold.Amount = order.FilledAmount
		s.hold.Price = order.AvgPrice
	} else {
		if ob.AskPrice() > s.hold.Price+s.StopWin {
			_, err = s.Buy(s.hold.Amount)
			s.hold.Amount = 0
			s.hold.Price = 0
			s.gear = 0
		} else if ob.AskPrice() < s.hold.Price-s.StopLoss && s.gear < s.MaxGear {
			_, err = s.Buy(s.hold.Amount)
			amount := s.hold.Amount * 2
			var addOrder *Order
			addOrder, err = s.Sell(amount)
			if err != nil {
				return
			}
			s.hold.Price = addOrder.Price
			s.hold.Amount = addOrder.FilledAmount
			s.gear++
		}
	}
	return nil
}

func (s *GamblerStrategy) Run() error {
	// run loop
	for {
		s.OnTick()
		fmt.Printf("加倍下注次数：%v 当前持仓：%#v", s.gear, s.hold)
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}

func (s *GamblerStrategy) OnExit() error {
	return nil
}

func (s *GamblerStrategy) Buy(amount float64) (order *Order, err error) {
	order, err = s.Exchange.OpenLong(s.Symbol,
		OrderTypeMarket, 0, amount)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("Order: %#v", order)
	return
}

func (s *GamblerStrategy) Sell(amount float64) (order *Order, err error) {
	order, err = s.Exchange.OpenShort(s.Symbol,
		OrderTypeMarket, 0, amount)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("Order: %#v", order)
	return
}

func main() {
	s := &GamblerStrategy{}
	err := serve.Serve(s)
	if err != nil {
		log.Printf("%v", err)
	}
}
