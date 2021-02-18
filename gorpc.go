package main

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"gorpc/factorycontract"
	"gorpc/paircontract"
	"gorpc/tokencontract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
)

func floatify(intString string, precision int) (float64, error) {
	if len(intString) <= precision {
		toAdd := precision - len(intString) + 1
		for i := 0; i < toAdd; i++ {
			intString = "0" + intString
		}
	}
	index := len(intString) - precision
	q := intString[:index] + "." + intString[index:]
	f, ok := new(big.Float).SetString(q)
	if !ok {
		return 0, errors.New("!!floatify(): Could not parse this resulting string to float: " + q)
	}
	res, _ := f.Float64()
	return res, nil
}

func getPairAddress(client *ethclient.Client, tokenA common.Address, tokenB common.Address) (common.Address, error) {
	cFactory, err := factorycontract.NewFactorycontract(common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"), client)
	if err != nil {
		return tokenA, err
	}
	pairAddress, err := cFactory.GetPair(nil, tokenA, tokenB)
	if err != nil {
		return tokenA, err
	}
	if pairAddress.Hex() == common.HexToAddress("0x0000000000000000000000000000000000000000").Hex() {
		return tokenA, errors.New("!!getPairAddress(): Pair does not exist")
	}
	return pairAddress, nil
}

func getDecimals(client *ethclient.Client, address common.Address) (int64, error) {
	cTokenA, err := tokencontract.NewTokencontract(address, client)
	if err != nil {
		return 0, err
	}
	decimals, err := cTokenA.Decimals(nil)
	return decimals.Int64(), err
}

func splitAddresses(addresses string) (common.Address, common.Address, error) {
	strs := strings.Split(addresses, "_")
	if len(strs) != 2 {
		return *new(common.Address), *new(common.Address), errors.New("!!splitAddresses(\"" + addresses + "\"): invalid input string")
	}
	return common.HexToAddress(strs[0]), common.HexToAddress(strs[1]), nil
}

func getPrice(addresses string, precision int64) (float64, error) {
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/a6995174790a4293904918f4ff7056de")
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in client, err := ethclient.Dial(\"https://mainnet.infura.io/v3/a6995174790a4293904918f4ff7056de\"): %s", err)
	}

	tokenA, tokenB, err := splitAddresses(addresses)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in tokenA, tokenB, err := splitAddresses(addresses): %s", err)
	}
	pairAddress, err := getPairAddress(client, tokenA, tokenB)
	if err != nil {
		return 0.0, fmt.Errorf("Error in getPrice in pairAddress, err := getPairAddress(client, tokenA, tokenB): %s", err)
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
