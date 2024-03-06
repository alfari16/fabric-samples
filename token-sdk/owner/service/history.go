/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package service

import (
	"time"

	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/ttx"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/ttxdb"
	"github.com/pkg/errors"
)

// SERVICE

type TransactionHistoryItem struct {
	// TxID is the transaction ID
	TxID string
	// ActionType is the type of action performed by this transaction record
	ActionType int
	// SenderEID is the enrollment ID of the account that is sending tokens
	Sender string
	// RecipientEID is the enrollment ID of the account that is receiving tokens
	Recipient string
	// TokenType is the type of token
	TokenType string
	// Amount is positive if tokens are received. Negative otherwise
	Amount int64

	TransactionType     string
	TransactionCategory string
	Platform            string

	// Timestamp is the time the transaction was submitted to the db
	Timestamp time.Time
	// Status is the status of the transaction
	Status string
	// Message is the user message sent with the transaction. It comes from
	// the ApplicationMetadata and is sent in the transient field
	Message string
}

func (s TokenService) GetHistoyByCred(fscName string, cred string, id string, param string, paramVal string) (walletId string, txs []TransactionHistoryItem, err error) {
	if id != "" {
		txs, err = s.GetHistory(id, param, paramVal)
		if err != nil {
			return "", txs, errors.Wrap(err, "failed to get transaction history")
		}

		return
	}

	signerData, err := s.getSignerDataByCred(fscName, cred)
	if err != nil {
		return "", txs, errors.Wrap(err, "failed to get signer data")
	}

	if signerData == nil {
		return "", txs, errors.New("signer data not found")
	}

	walletId = signerData.EnrollmentID

	txs, err = s.GetHistory(id, param, paramVal)
	if err != nil {
		return "", txs, errors.Wrap(err, "failed to get transaction history")
	}

	return

}

// GetHistory returns the full transaction history for an owner.
func (s TokenService) GetHistory(wallet string, param string, paramVal string) (txs []TransactionHistoryItem, err error) {
	// Get query executor
	owner := ttx.NewOwner(s.FSC, token.GetManagementService(s.FSC))
	aqe := owner.NewQueryExecutor()
	defer aqe.Done()
	it, err := aqe.Transactions(ttxdb.QueryTransactionsParams{
		SenderWallet:    wallet,
		RecipientWallet: wallet,
	})
	if err != nil {
		return txs, errors.Wrap(err, "failed querying transactions from db")
	}

	defer it.Close()

	// we need transaction info to get the transient field (application metadata)
	tip := ttx.NewTransactionInfoProvider(s.FSC, token.GetManagementService(s.FSC))
	if tip == nil {
		return txs, errors.New("failed to get transactionInfoProvider")
	}

	// Return the list of accepted transactions
	for {
		tx, err := it.Next()
		if err != nil {
			return txs, errors.Wrap(err, "failed iterating over transactions")
		}
		if tx == nil {
			break
		}
		transaction := TransactionHistoryItem{
			TxID:       tx.TxID,
			ActionType: int(tx.ActionType),
			Sender:     tx.SenderEID,
			Recipient:  tx.RecipientEID,
			TokenType:  tx.TokenType,
			Amount:     tx.Amount.Int64(),
			Timestamp:  tx.Timestamp.UTC(),
			Status:     string(tx.Status),
		}
		// set user provided message from transient field
		ti, err := tip.TransactionInfo(tx.TxID)
		if err != nil {
			return txs, errors.Wrapf(err, "cannot get transaction info for %s", tx.TxID)
		}
		if ti.ApplicationMetadata != nil {
			if string(ti.ApplicationMetadata["message"]) != "" {
				transaction.Message = string(ti.ApplicationMetadata["message"])
			}

			if string(ti.ApplicationMetadata["trx_type"]) != "" {
				transaction.TransactionType = string(ti.ApplicationMetadata["trx_type"])
			}

			if string(ti.ApplicationMetadata["trx_category"]) != "" {
				transaction.TransactionCategory = string(ti.ApplicationMetadata["trx_category"])
			}

			if string(ti.ApplicationMetadata["platform"]) != "" {
				transaction.Platform = string(ti.ApplicationMetadata["platform"])
			}

		}
		txs = append(txs, transaction)
	}

	txs, err = filterByParam(txs, param, paramVal)

	return
}

func filterByParam(txs []TransactionHistoryItem, param string, paramVal string) (filteredTxs []TransactionHistoryItem, err error) {
	switch param {
	case "":
		filteredTxs = txs
		err = nil
	case "transactionType":
		for _, tx := range txs {
			if tx.TransactionType == paramVal {
				filteredTxs = append(filteredTxs, tx)
			}
		}
		err = nil
	case "transactionCategory":
		for _, tx := range txs {
			if tx.TransactionCategory == paramVal {
				filteredTxs = append(filteredTxs, tx)
			}
		}
		err = nil
	default:
		filteredTxs = txs
		err = errors.New("invalid filter parameter")
	}

	return
}
