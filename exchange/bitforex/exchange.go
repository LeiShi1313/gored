package bitforex

// Copyright (c) 2015-2019 Bitontop Technologies Inc.
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map"

	"github.com/bitontop/gored/coin"
	"github.com/bitontop/gored/exchange"
	"github.com/bitontop/gored/pair"
	"github.com/bitontop/gored/utils"
)

type Bitforex struct {
	ID      int
	Name    string `bson:"name"`
	Website string `bson:"website"`

	API_KEY    string
	API_SECRET string

	Source    exchange.DataSource // / exchange API / microservicve api 1 / PSQL
	SourceURI string
}

var pairConstraintMap cmap.ConcurrentMap
var coinConstraintMap cmap.ConcurrentMap
var balanceMap cmap.ConcurrentMap

var instance *Bitforex
var once sync.Once

/***************************************************/
func CreateBitforex(config *exchange.Config) *Bitforex {
	once.Do(func() {
		instance = &Bitforex{
			ID:      DEFAULT_ID,
			Name:    "Bitforex",
			Website: "https://api.bitforex.com",

			API_KEY:    config.API_KEY,
			API_SECRET: config.API_SECRET,
			Source:     config.Source,
			SourceURI:  config.SourceURI,
		}

		balanceMap = cmap.New()
		coinConstraintMap = cmap.New()
		pairConstraintMap = cmap.New()

		instance.InitData()
	})
	return instance
}

func (e *Bitforex) InitData() {
	switch e.Source {
	case exchange.EXCHANGE_API:
		e.GetCoinsData()
		time.Sleep(time.Second * 10)
		e.GetPairsData()
		break
	case exchange.MICROSERVICE_API:
		break
	case exchange.JSON_FILE:
		exchangeData := utils.GetExchangeDataFromJSON(e.SourceURI, e.GetName())
		coinConstraintMap = exchangeData.CoinConstraint
		pairConstraintMap = exchangeData.PairConstraint
		break
	case exchange.PSQL:
	default:
		log.Printf("Bitforex Initial Coin: There is not selected data source.")
	}
}

/**************** Exchange Information ****************/
func (e *Bitforex) GetID() int {
	return e.ID
}

func (e *Bitforex) GetName() exchange.ExchangeName {
	return exchange.BITFOREX
}

func (e *Bitforex) GetBalance(coin *coin.Coin) float64 {
	if tmp, ok := balanceMap.Get(coin.Code); ok {
		return tmp.(float64)
	} else {
		return 0.0
	}
}

func (e *Bitforex) GetTradingWebURL(pair *pair.Pair) string {
	return fmt.Sprintf("https://www.bitforex.com/cn/trade/spotTrading?commodityCode=%s&currencyCode=%s", e.GetSymbolByCoin(pair.Target), e.GetSymbolByCoin(pair.Base))
}

/*************** Coins on the Exchanges ***************/
func (e *Bitforex) GetCoinConstraint(coin *coin.Coin) *exchange.CoinConstraint {
	if tmp, ok := coinConstraintMap.Get(fmt.Sprintf("%d", coin.ID)); ok {
		return tmp.(*exchange.CoinConstraint)
	}
	return nil
}

func (e *Bitforex) SetCoinConstraint(coinConstraint *exchange.CoinConstraint) {
	coinConstraintMap.Set(fmt.Sprintf("%d", coinConstraint.CoinID), coinConstraint)
}

func (e *Bitforex) GetCoins() []*coin.Coin {
	coinList := []*coin.Coin{}
	keySort := []int{}
	for _, key := range coinConstraintMap.Keys() {
		id, _ := strconv.Atoi(key)
		keySort = append(keySort, id)
	}
	sort.Ints(keySort)
	for _, key := range keySort {
		c := coin.GetCoinByID(key)
		if c != nil {
			coinList = append(coinList, c)
		}
	}
	return coinList
}

func (e *Bitforex) GetSymbolByCoin(coin *coin.Coin) string {
	key := fmt.Sprintf("%d", coin.ID)
	if tmp, ok := coinConstraintMap.Get(key); ok {
		cc := tmp.(*exchange.CoinConstraint)
		return cc.ExSymbol
	}
	return ""
}

func (e *Bitforex) GetCoinBySymbol(symbol string) *coin.Coin {
	for _, id := range coinConstraintMap.Keys() {
		if tmp, ok := coinConstraintMap.Get(id); ok {
			cc := tmp.(*exchange.CoinConstraint)
			if cc.ExSymbol == symbol {
				return cc.Coin
			}
		}
	}
	return nil
}

func (e *Bitforex) DeleteCoin(coin *coin.Coin) {
	coinConstraintMap.Remove(fmt.Sprintf("%d", coin.ID))
}

/*************** Pairs on the Exchanges ***************/
func (e *Bitforex) GetPairConstraint(pair *pair.Pair) *exchange.PairConstraint {
	if tmp, ok := pairConstraintMap.Get(fmt.Sprintf("%d", pair.ID)); ok {
		return tmp.(*exchange.PairConstraint)
	}
	return nil
}

func (e *Bitforex) SetPairConstraint(pairConstraint *exchange.PairConstraint) {
	pairConstraintMap.Set(fmt.Sprintf("%d", pairConstraint.PairID), pairConstraint)
}

func (e *Bitforex) GetPairs() []*pair.Pair {
	pairList := []*pair.Pair{}
	keySort := []int{}
	for _, key := range pairConstraintMap.Keys() {
		id, _ := strconv.Atoi(key)
		keySort = append(keySort, id)
	}
	sort.Ints(keySort)
	for _, key := range keySort {
		p := pair.GetPairByID(key)
		if p != nil {
			pairList = append(pairList, p)
		}
	}
	return pairList
}

func (e *Bitforex) GetPairBySymbol(symbol string) *pair.Pair {
	for _, id := range pairConstraintMap.Keys() {
		if tmp, ok := pairConstraintMap.Get(id); ok {
			pc := tmp.(*exchange.PairConstraint)
			if pc.ExSymbol == symbol {
				return pc.Pair
			}
		}
	}
	return nil
}

func (e *Bitforex) GetSymbolByPair(pair *pair.Pair) string {
	pairConstraint := e.GetPairConstraint(pair)
	if pairConstraint != nil {
		return pairConstraint.ExSymbol
	}
	return ""
}

func (e *Bitforex) HasPair(pair *pair.Pair) bool {
	return pairConstraintMap.Has(fmt.Sprintf("%d", pair.ID))
}

func (e *Bitforex) DeletePair(pair *pair.Pair) {
	pairConstraintMap.Remove(fmt.Sprintf("%d", pair.ID))
}

/**************** Exchange Constraint ****************/
func (e *Bitforex) GetConstraintFetchMethod(pair *pair.Pair) *exchange.ConstrainFetchMethod {
	constrainFetchMethod := &exchange.ConstrainFetchMethod{}
	constrainFetchMethod.Fee = false
	constrainFetchMethod.LotSize = true
	constrainFetchMethod.PriceFilter = true
	constrainFetchMethod.TxFee = false
	constrainFetchMethod.Withdraw = false
	constrainFetchMethod.Deposit = false
	constrainFetchMethod.Confirmation = false
	return constrainFetchMethod
}

func (e *Bitforex) UpdateConstraint() {
	e.GetCoinsData()
	e.GetPairsData()
}

/**************** Coin Constraint ****************/
func (e *Bitforex) GetTxFee(coin *coin.Coin) float64 {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.TxFee
}

func (e *Bitforex) CanWithdraw(coin *coin.Coin) bool {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Withdraw
}

func (e *Bitforex) CanDeposit(coin *coin.Coin) bool {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Deposit
}

func (e *Bitforex) GetConfirmation(coin *coin.Coin) int {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Confirmation
}

/**************** Pair Constraint ****************/
func (e *Bitforex) GetFee(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.TakerFee
}

func (e *Bitforex) GetLotSize(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.LotSize
}

func (e *Bitforex) GetPriceFilter(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.PriceFilter
}