package database

import "github.com/segmentio/ksuid"

type IDPrefix string

const (
	PrefixAccount        IDPrefix = "acct"
	PrefixSecurity       IDPrefix = "sec"
	PrefixPosition       IDPrefix = "pos"
	PrefixTransaction    IDPrefix = "txn"
	PrefixLot            IDPrefix = "lot"
	PrefixLotDisposition IDPrefix = "disp"
	PrefixCashTxn        IDPrefix = "cash"
)

func NewID(prefix IDPrefix) string {
	return string(prefix) + "_" + ksuid.New().String()
}
