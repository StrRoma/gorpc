package main

import (
	"fmt"
	"math/big"

	"gorpc/paircontract"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getPriceCake(addresses string, precision int64) (float64, error) {
	client, err := ethclient.Dial("https://bsc-dataseed.binance.org/")
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in client, err := ethclient.Dial(\"https://bsc-dataseed.binance.org/\"): %s", err)
	}

	tokenA, tokenB, err := splitAddresses(addresses)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in tokenA, tokenB, err := splitAddresses(addresses): %s", err)
	}
	pairAddress, err := getPairAddressCake(client, tokenA, tokenB)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in pairAddress, err := getPairAddressCake(client, tokenA, tokenB): %s", err)
	}
	cPair, err := paircontract.NewPaircontract(pairAddress, client)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in cPair, err := paircontract.NewPaircontract(pairAddress, client): %s", err)
	}
	decimalsA, err := getDecimals(client, tokenA)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in decimalsA, err := getDecimals(client, tokenA): %s", err)
	}
	decimalsB, err := getDecimals(client, tokenB)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in decimalsB, err := getDecimals(client, tokenB): %s", err)
	}
	reserves, err := cPair.GetReserves(nil)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in reserves, err := cPair.GetReserves(nil): %s", err)
	}
	token0, err := cPair.Token0(nil)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in token0, err := cPair.Token0(nil): %s", err)
	}
	if token0.Hex() == tokenA.Hex() {
		result, ok := new(big.Int).SetString(reserves.Reserve1.String(), 10)
		if !ok {
			return 0.0, fmt.Errorf("!!SetString(): conversion problem")
		}
		result.Mul(result, math.BigPow(10, decimalsA+precision)).Div(result, math.BigPow(10, decimalsB)).Div(result, reserves.Reserve0)
		return floatify(result.String(), int(precision))
	}
	result, ok := new(big.Int).SetString(reserves.Reserve0.String(), 10)
	if !ok {
		return 0.0, fmt.Errorf("!!SetString(): conversion problem")
	}
	result.Mul(result, math.BigPow(10, decimalsA+precision)).Div(result, math.BigPow(10, decimalsB)).Div(result, reserves.Reserve1)
	return floatify(result.String(), int(precision))
}
