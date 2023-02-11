package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

func handlePayments(w http.ResponseWriter, r *http.Request, re *Relay) {
	w.Header().Set("Content-Type", "application/json")

	type paymentRequest struct {
		TxHash string `json:"tx_hash"`
		PubKey string `json:"pubkey"`
	}

	var req paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to decode request body: " + err.Error(),
		})
		return
	}

	txHash := common.HexToHash(req.TxHash)
	var tx *types.Transaction
	tx, isPending, err := re.client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to get transaction: " + err.Error(),
		})
		return
	}
	if isPending {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "tx is still pending",
		})
		return
	}

	if re.amount.Cmp(tx.Value()) != 1 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "insufficient amount",
		})
		return
	}

	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(req.PubKey)[1:])
	tx_receiver := hexutil.Encode(hash.Sum(nil)[12:])
	if tx_receiver != GetTransactionMessage(tx).From().Hex() {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "pubkey's hex is not match with tx's hex",
		})
		return
	}

	if re.address != tx.To().Hex() {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "tx send to wrong address",
		})
		return
	}

	err = re.storage.SavePayment(req.PubKey, req.TxHash)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unable to save payment: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(tx)
}

func GetTransactionMessage(tx *types.Transaction) types.Message {
	msg, err := tx.AsMessage(types.LatestSignerForChainID(tx.ChainId()), nil)
	if err != nil {
		log.Fatal(err)
	}
	return msg
}
