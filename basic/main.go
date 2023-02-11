package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fiatjaf/relayer"
	"github.com/fiatjaf/relayer/basic/pgsql"
	"github.com/fiatjaf/relayer/storage/postgresql"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr"
)

type Relay struct {
	PostgresDatabase string `envconfig:"POSTGRESQL_DATABASE"`

	storage *pgsql.BasicPostgresBackend

	client *ethclient.Client

	amount *big.Int

	address string
}

func NewRelay(client *ethclient.Client, amount *big.Int, address string) Relay {
	return Relay{
		client:  client,
		amount:  amount,
		address: address,
	}
}

func (r *Relay) Name() string {
	return "BasicRelay"
}

func (r *Relay) Storage() relayer.Storage {
	return r.storage
}

func (r *Relay) OnInitialized(s *relayer.Server) {
	s.Router().Path("/payments").Methods("POST").Headers("Accept", "application/json").HandlerFunc(func(w http.ResponseWriter, rq *http.Request) { handlePayments(w, rq, r) })
}

func (r *Relay) Init() error {
	err := envconfig.Process("", r)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}

	// every hour, delete all very old events
	go func() {
		db := r.Storage().(*pgsql.BasicPostgresBackend)

		for {
			time.Sleep(60 * time.Minute)
			db.DB.Exec(`DELETE FROM event WHERE created_at < $1`, time.Now().AddDate(0, -3, 0).Unix()) // 3 months
		}
	}()

	return nil
}

func (r *Relay) AcceptEvent(evt *nostr.Event) bool {
	// only accept they have a good preimage for a paid invoice for their public key
	isPaid, err := r.storage.CheckPayment(evt.PubKey)
	if err != nil {
		log.Printf("unable to fetch payment for accept event: %s", err.Error())
	}
	if !isPaid {
		return false
	}

	// block events that are too large
	jsonb, _ := json.Marshal(evt)
	if len(jsonb) > 10000 {
		return false
	}

	return true
}

func main() {
	chainName := os.Getenv("CHAIN_NAME")
	paymentAmount := os.Getenv("PAYMENT_AMOUNT")
	paymentAddress := os.Getenv("PAYMENT_ADDRESS")
	if chainName == "" || paymentAmount == "" || paymentAddress == "" {
		log.Fatal("env is not set")
	}

	cl, err := GetEthClient(chainName)
	if err != nil {
		log.Fatalf("init eth cline filed: %v", err)
	}
	r := NewRelay(cl, ToWei(paymentAmount, 18), paymentAddress)
	if err := envconfig.Process("", &r); err != nil {
		log.Fatalf("failed to read from env: %v", err)
		return
	}
	db := &pgsql.BasicPostgresBackend{
		PostgresBackend: &postgresql.PostgresBackend{DatabaseURL: r.PostgresDatabase},
	}
	if err := db.PaymentInit(); err != nil {
		log.Fatalf("fil to init payment table: %v", err)
	}
	r.storage = db
	if err := relayer.Start(&r); err != nil {
		log.Fatalf("server terminated: %v", err)
	}
}
