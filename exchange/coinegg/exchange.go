package binance

// Copyright (c) 2015-2019 Bitontop Technologies Inc.
// Distributed under the MIT software license, see the accompanying
// file COPYING or http://www.opensource.org/licenses/mit-license.php.

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"

	cmap "github.com/orcaman/concurrent-map"

	"../../coin"
	"../../exchange"
	"../../pair"
	"../../utils"
)

type Binance struct {
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

var instance *Binance
var once sync.Once

/***************************************************/
func CreateBinance(config *exchange.Config) *Binance {
	once.Do(func() {
		instance = &Binance{
			ID:      DEFAULT_ID,
			Name:    "Binance",
			Website: "https://www.binance.com/",

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

func (e *Binance) InitData() {
	switch e.Source {
	case exchange.EXCHANGE_API:
		e.GetCoinsData()
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
		log.Printf("Binance Initial Coin: There is not selected data source.")
	}
}

/**************** Exchange Information ****************/
func (e *Binance) GetID() int {
	return e.ID
}

func (e *Binance) GetName() exchange.ExchangeName {
	return exchange.BINANCE
}

func (e *Binance) GetTradingWebURL(pair *pair.Pair) string {
	return fmt.Sprintf("https://www.binance.com/cn/trade/pro/%s_%s", e.GetSymbolByCoin(pair.Target), e.GetSymbolByCoin(pair.Base))
}

func (e *Binance) GetBalance(coin *coin.Coin) float64 {
	if tmp, ok := balanceMap.Get(coin.Code); ok {
		return tmp.(float64)
	} else {
		return 0.0
	}
}

/*************** Coins on the Exchanges ***************/
func (e *Binance) GetCoinConstraint(coin *coin.Coin) *exchange.CoinConstraint {
	if tmp, ok := coinConstraintMap.Get(fmt.Sprintf("%d", coin.ID)); ok {
		return tmp.(*exchange.CoinConstraint)
	}
	return nil
}

func (e *Binance) SetCoinConstraint(coinConstraint *exchange.CoinConstraint) {
	coinConstraintMap.Set(fmt.Sprintf("%d", coinConstraint.CoinID), coinConstraint)
}

func (e *Binance) GetCoins() []*coin.Coin {
	coinList := []*coin.Coin{}
	keySort := []int{}
	for _, key := range coinConstraintMap.Keys() {
		id, _ := strconv.Atoi(key)
		keySort = append(keySort, id)
	}
	sort.Ints(keySort)
	for _, key := range keySort {
		coinList = append(coinList, coin.GetCoinByID(key))
	}
	return coinList
}

func (e *Binance) GetCoinBySymbol(symbol string) *coin.Coin {
	for _, id := range coinConstraintMap.Keys() {
		if tmp, ok := coinConstraintMap.Get(id); ok {
			cc := tmp.(*exchange.CoinConstraint)
			if cc.ExSymbol == symbol {
				return cc.Coin
			}
		} else {
			log.Printf("Get ID %s CoinConstraint Err", id)
		}
	}
	return nil
}

func (e *Binance) GetSymbolByCoin(coin *coin.Coin) string {
	key := fmt.Sprintf("%d", coin.ID)
	if tmp, ok := coinConstraintMap.Get(key); ok {
		cc := tmp.(*exchange.CoinConstraint)
		return cc.ExSymbol
	}
	return ""
}

func (e *Binance) DeleteCoin(coin *coin.Coin) {
	coinConstraintMap.Remove(fmt.Sprintf("%d", coin.ID))
}

/*************** Pairs on the Exchanges ***************/
func (e *Binance) GetPairConstraint(pair *pair.Pair) *exchange.PairConstraint {
	if tmp, ok := pairConstraintMap.Get(fmt.Sprintf("%d", pair.ID)); ok {
		return tmp.(*exchange.PairConstraint)
	}
	return nil
}

func (e *Binance) SetPairConstraint(pairConstraint *exchange.PairConstraint) {
	pairConstraintMap.Set(fmt.Sprintf("%d", pairConstraint.PairID), pairConstraint)
}

func (e *Binance) GetPairs() []*pair.Pair {
	pairList := []*pair.Pair{}
	keySort := []int{}
	for _, key := range pairConstraintMap.Keys() {
		id, _ := strconv.Atoi(key)
		keySort = append(keySort, id)
	}
	sort.Ints(keySort)
	for _, key := range keySort {
		pairList = append(pairList, pair.GetPairByID(key))
	}
	return pairList
}

func (e *Binance) GetPairBySymbol(symbol string) *pair.Pair {
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

func (e *Binance) GetSymbolByPair(pair *pair.Pair) string {
	pairConstraint := e.GetPairConstraint(pair)
	if pairConstraint != nil {
		return pairConstraint.ExSymbol
	}
	return ""
}

func (e *Binance) HasPair(pair *pair.Pair) bool {
	return pairConstraintMap.Has(fmt.Sprintf("%d", pair.ID))
}

func (e *Binance) DeletePair(pair *pair.Pair) {
	pairConstraintMap.Remove(fmt.Sprintf("%d", pair.ID))
}

/**************** Exchange Constraint ****************/
func (e *Binance) GetConstraintFetchMethod(pair *pair.Pair) *exchange.ConstrainFetchMethod {
	constrainFetchMethod := &exchange.ConstrainFetchMethod{}
	constrainFetchMethod.Fee = true
	constrainFetchMethod.LotSize = true
	constrainFetchMethod.PriceFilter = true
	constrainFetchMethod.TxFee = true
	constrainFetchMethod.Withdraw = true
	constrainFetchMethod.Deposit = true
	constrainFetchMethod.Confirmation = true
	return constrainFetchMethod
}

func (e *Binance) UpdateConstraint() {
	e.GetCoinsData()
	e.GetPairsData()
}

/**************** Coin Constraint ****************/
func (e *Binance) GetTxFee(coin *coin.Coin) float64 {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.TxFee
}

func (e *Binance) CanWithdraw(coin *coin.Coin) bool {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Withdraw
}

func (e *Binance) CanDeposit(coin *coin.Coin) bool {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Deposit
}

func (e *Binance) GetConfirmation(coin *coin.Coin) int {
	coinConstraint := e.GetCoinConstraint(coin)
	return coinConstraint.Confirmation
}

/**************** Pair Constraint ****************/
func (e *Binance) GetFee(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.TakerFee
}

func (e *Binance) GetLotSize(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.LotSize
}

func (e *Binance) GetPriceFilter(pair *pair.Pair) float64 {
	pairConstraint := e.GetPairConstraint(pair)
	return pairConstraint.PriceFilter
}
