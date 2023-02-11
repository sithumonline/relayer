relayer basic
=============

  - a basic relay implementation based on relayer.
  - uses postgres, which I think must be over version 12 since it uses generated columns.
  - it has some antispam limits, tries to delete old stuff so things don't get out of control, and some other small optimizations.

running
-------

grab a binary from the releases page and run it with the environment variable POSTGRESQL_DATABASE set to some postgres url:

    POSTGRESQL_DATABASE=postgres://name:pass@localhost:5432/dbname ./relayer-basic

it also accepts a HOST and a PORT environment variables.

```bash
PAYMENT_AMOUNT=19832284976715000 CHAIN_NAME="Ethereum Mainnet" PAYMENT_ADDRESS=0xa6B7C79E98E277153c7f237c347aD3Bb3819Fef6 go run .
```

compiling
---------

if you know Go you already know this:

    go install github.com/fiatjaf/relayer/basic

or something like that.

payment
-------

```bash
curl --location --request POST 'http://0.0.0.0:7447/payments' \
--header 'Accept: application/json' \
--header 'Content-Type: application/json' \
--data-raw '{
    "tx_hash": "0x4e76c96f6f2d41b2b9d3324273c6d903ab5f800ea7e57e2e9d3ce60873bf6163",
    "pvtkey": "4af55b0ff75986ec3294f9d6fbf860c0b5df77984ad542adf9abb186462819d8"   
}
'
```
