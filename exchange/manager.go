package exchange

// Copyright (c) 2015-2019 Bitontop Technologies Inc.
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

import (
	"log"
	"sync"
	"time"

	"../coin"
	"../pair"

	cmap "github.com/orcaman/concurrent-map"
)

type Exchange interface {
	InitData()

	/***** Exchange Information *****/
	GetID() int
	GetName() ExchangeName
	GetTradingWebURL(pair *pair.Pair) string
	GetBalance(coin *coin.Coin) float64

	/***** Coin Information *****/
	GetCoinConstraint(coin *coin.Coin) *CoinConstraint
	SetCoinConstraint(coinConstraint *CoinConstraint)
	GetCoins() []*coin.Coin
	GetCoinBySymbol(symbol string) *coin.Coin
	GetSymbolByCoin(coin *coin.Coin) string
	DeleteCoin(coin *coin.Coin)

	/***** Pair Information *****/
	GetPairConstraint(pair *pair.Pair) *PairConstraint
	SetPairConstraint(pairConstraint *PairConstraint)
	GetPairs() []*pair.Pair
	GetPairBySymbol(symbol string) *pair.Pair
	GetSymbolByPair(pair *pair.Pair) string
	HasPair(*pair.Pair) bool
	DeletePair(pair *pair.Pair)

	/***** Public API *****/
	GetCoinsData()
	GetPairsData()
	OrderBook(p *pair.Pair) (*Maker, error)

	/***** Private API *****/
	UpdateAllBalances()
	Withdraw(coin *coin.Coin, quantity float64, addr, tag string) bool

	LimitSell(pair *pair.Pair, quantity, rate float64) (*Order, error)
	LimitBuy(pair *pair.Pair, quantity, rate float64) (*Order, error)

	OrderStatus(order *Order) error
	ListOrders() ([]*Order, error)

	CancelOrder(order *Order) error
	CancelAllOrder() error //TODO need to impl cancel all order for exchanges

	/***** Exchange Constraint *****/
	GetConstraintFetchMethod(pair *pair.Pair) *ConstrainFetchMethod
	UpdateConstraint()
	/***** Coin Constraint *****/
	GetTxFee(coin *coin.Coin) float64
	CanWithdraw(coin *coin.Coin) bool
	CanDeposit(coin *coin.Coin) bool
	GetConfirmation(coin *coin.Coin) int
	/***** Pair Constraint *****/
	GetFee(pair *pair.Pair) float64
	GetLotSize(pair *pair.Pair) float64
	GetPriceFilter(pair *pair.Pair) float64
}

type ExchangeManager struct {
}

var instance *ExchangeManager
var once sync.Once

var exMap cmap.ConcurrentMap
var exList = make([]Exchange, 0)
var supportList = make([]ExchangeName, 0)

func CreateExchangeManager() *ExchangeManager {
	once.Do(func() {
		instance = &ExchangeManager{}
		instance.init()

		if exMap == nil {
			exMap = cmap.New()
		}
	})
	return instance
}

func (e *ExchangeManager) init() {
	e.initExchangeNames()
}

func (e *ExchangeManager) Add(exchange Exchange) {
	key := string(exchange.GetName())
	exMap.Set(key, exchange)
}

func (e *ExchangeManager) Get(name ExchangeName) Exchange {
	if tmp, ok := exMap.Get(string(name)); ok {
		return tmp.(Exchange)
	}
	return nil
}

func (e *ExchangeManager) GetID(name ExchangeName) int {
	eInstance := e.Get(name)
	return eInstance.GetID()
}

func (e *ExchangeManager) GetById(i int) Exchange {
	for _, ex := range e.GetExchanges() {
		if ex.GetID() == i {
			return ex
		}
	}
	return nil
}

func (e *ExchangeManager) GetStr(name string) Exchange {
	for _, v := range e.GetSupportExchanges() {
		if string(v) == name {
			return e.Get(v)
		}
	}
	return nil
}

func (e *ExchangeManager) GetSupportExchanges() []ExchangeName {
	return supportList
}

func (e *ExchangeManager) GetExchanges() []Exchange {
	exchanges := []Exchange{}
	for _, key := range exMap.Keys() {
		exchanges = append(exchanges, e.GetStr(key))
	}

	return exchanges
}

func (e *ExchangeManager) UpdateExData(conf *Update) {
	switch conf.Method {
	case API_TIGGER:
		break
	case TIME_TIGGER:
		for {
			for _, exName := range conf.ExNames {
				eInstance := e.Get(exName)
				eInstance.InitData()
				log.Printf("%s Data Updated. Coin: %d   Pair: %d", exName, len(eInstance.GetCoins()), len(eInstance.GetPairs()))
			}
			time.Sleep(conf.Time)
		}
	}
}
