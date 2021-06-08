package main

import (
	"errors"
	"gorpc/factorycontract"
	"gorpc/tokencontract"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
