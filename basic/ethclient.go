package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

type Chain struct {
	Name     string   `json:"name,omitempty"`
	Chain    string   `json:"chain,omitempty"`
	Icon     string   `json:"icon,omitempty"`
	RPC      []string `json:"rpc,omitempty"`
	Features []struct {
		Name string `json:"name,omitempty"`
	} `json:"features,omitempty"`
	Faucets        []interface{} `json:"faucets,omitempty"`
	NativeCurrency struct {
		Name     string `json:"name,omitempty"`
		Symbol   string `json:"symbol,omitempty"`
		Decimals int    `json:"decimals,omitempty"`
	} `json:"nativeCurrency,omitempty"`
	InfoURL   string `json:"infoURL,omitempty"`
	ShortName string `json:"shortName,omitempty"`
	ChainID   int    `json:"chainId,omitempty"`
	NetworkID int    `json:"networkId,omitempty"`
	Slip44    int    `json:"slip44,omitempty"`
	Ens       struct {
		Registry string `json:"registry,omitempty"`
	} `json:"ens,omitempty"`
	Explorers []struct {
		Name     string `json:"name,omitempty"`
		URL      string `json:"url,omitempty"`
		Standard string `json:"standard,omitempty"`
	} `json:"explorers,omitempty"`
}

func getChains() ([]Chain, error) {
	resp, err := http.Get("https://chainid.network/chains.json")
	if err != nil {
		return nil, fmt.Errorf("error unable to fetch chains.json: %w", err)
	}

	defer resp.Body.Close()
	var chains []Chain
	if err := json.NewDecoder(resp.Body).Decode(&chains); err != nil {
		return nil, fmt.Errorf("error unable to decode chains.json: %w", err)
	}

	return chains, nil
}

func getRpcUrl(name string) (string, error) {
	chains, err := getChains()
	if err != nil {
		return "", fmt.Errorf("error unable to get chains: %w", err)
	}

	var RPCs []string
	for _, chain := range chains {
		if chain.Name != name {
			continue
		}

		const INFURA_API_KEY = "INFURA_API_KEY"
		INFURA_API_KEY_ENV := os.Getenv(INFURA_API_KEY)

		if INFURA_API_KEY_ENV == "" {
			for _, URL := range chain.RPC {
				if strings.Contains(URL, INFURA_API_KEY) {
					continue
				}

				RPCs = append(RPCs, URL)
			}
		} else {
			for _, URL := range chain.RPC {
				if strings.Contains(URL, INFURA_API_KEY) {
					URL = strings.Replace(URL, INFURA_API_KEY, INFURA_API_KEY_ENV, -1)
				}

				RPCs = append(RPCs, URL)
			}
		}

		break
	}

	if len(RPCs) == 0 {
		return "", fmt.Errorf("RPC URL is not found for %s", name)
	}

	rand.Seed(time.Now().UnixNano())
	return RPCs[rand.Intn(len(RPCs))], nil
}

func GetEthClient(name string) (*ethclient.Client, error) {
	url, err := getRpcUrl(name)
	if err != nil {
		return nil, fmt.Errorf("error unable to get RPC URL: %w", err)
	}
	log.Printf("RPC URL: %s\n", url)
	return ethclient.Dial(url)
}

func ToWei(iamount interface{}, decimals int) *big.Int {
	amount := decimal.NewFromFloat(0)
	switch v := iamount.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	wei := new(big.Int)
	wei.SetString(result.String(), 10)

	return wei
}
