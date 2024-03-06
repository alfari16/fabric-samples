/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-samples/token-sdk/owner/service"
)

type AppEnv struct {
	FSCName   string
	COmpanyID string
}

type Controller struct {
	Service service.TokenService
	AppEnv  AppEnv
}

func (c Controller) RegisterUser(ctx context.Context, request RegisterUserRequestObject) (RegisterUserResponseObject, error) {
	id := request.Body.IdWallet
	username := request.Body.UsernameWallet

	path, err := c.Service.RegisterOwnerWallet(c.AppEnv.FSCName, id, username)
	if err != nil {
		return RegisterUserdefaultJSONResponse{
			Body: Error{
				Message: "can't register user",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return RegisterUser200JSONResponse{
		TransferSuccessJSONResponse: TransferSuccessJSONResponse{
			Message: fmt.Sprintf("registered user %s", username),
			Payload: path,
		},
	}, nil
}

// Transfer tokens to another account
// (POST /owner/accounts/{id}/transfer)
func (c Controller) Transfer(ctx context.Context, request TransferRequestObject) (TransferResponseObject, error) {
	code := request.Body.Amount.Code
	value := uint64(request.Body.Amount.Value)
	sender := request.Id
	recipient := request.Body.Counterparty.Account
	recipientNode := request.Body.Counterparty.Node
	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	if trxCat == "" || trxType == "" {
		return TransferdefaultJSONResponse{
			Body: Error{
				Message: "Can't transfer funds",
				Payload: "transaction_category and transaction_type are required",
			},
			StatusCode: 400,
		}, nil
	}

	txID, err := c.Service.TransferTokens(code, value, trxType, trxCat, platform, sender, recipient, recipientNode, message)
	if err != nil {
		return TransferdefaultJSONResponse{
			Body: Error{
				Message: "can't transfer funds",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}
	return Transfer200JSONResponse{
		TransferSuccessJSONResponse: TransferSuccessJSONResponse{
			Message: fmt.Sprintf("%s transferred %d %s to %s", sender, value, code, recipient),
			Payload: txID,
		},
	}, nil
}

// Transfer by cred me
// (POST /owner/accounts/transfer/me)
func (c Controller) TransferMe(ctx context.Context, request TransferMeRequestObject) (TransferMeResponseObject, error) {
	code := request.Body.Amount.Code
	value := uint64(request.Body.Amount.Value)
	// cred := request.Params.Cred
	recipient := request.Body.Counterparty.Account
	recipientNode := request.Body.Counterparty.Node
	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	if trxCat == "" || trxType == "" {
		return TransferMedefaultJSONResponse{
			Body: Error{
				Message: "Can't transfer funds",
				Payload: "transaction_category and transaction_type are required",
			},
			StatusCode: 400,
		}, nil
	}

	var cred string
	var id string
	if request.Params.Cred == nil && request.Params.WalletId == nil {
		return TransferMedefaultJSONResponse{
			Body: Error{
				Message: "can't get history",
				Payload: "one of the header 'wallet_id' or 'cred' is required",
			},
			StatusCode: 400,
		}, nil
	} else {
		if request.Params.Cred != nil {
			cred = *request.Params.Cred
		} else if request.Params.WalletId != nil {
			id = *request.Params.WalletId
		}
	}

	walId, txID, err := c.Service.TransferTokensByCred(c.AppEnv.FSCName, code, value, trxType, trxCat, platform, cred, id, recipient, recipientNode, message)
	if err != nil {
		return TransferMedefaultJSONResponse{
			Body: Error{
				Message: "can't transfer funds",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return TransferMe200JSONResponse{
		TransferSuccessJSONResponse: TransferSuccessJSONResponse{
			Message: fmt.Sprintf("%s transferred %d %s to %s", walId, value, code, recipient),
			Payload: txID,
		},
	}, nil
}

// Get all accounts on this node and their balances
// (GET /owner/accounts)
func (c Controller) OwnerAccounts(ctx context.Context, request OwnerAccountsRequestObject) (OwnerAccountsResponseObject, error) {
	balances, err := c.Service.GetAllBalances()
	if err != nil {
		return OwnerAccountsdefaultJSONResponse{
			Body: Error{
				Message: "can't get accounts",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	acc := []Account{}
	for wallet, balance := range balances {
		amounts := []Amount{}
		for typ, val := range balance {
			amounts = append(amounts, Amount{
				Code:  typ,
				Value: val,
			})
		}
		acc = append(acc, Account{
			Id:      wallet,
			Balance: amounts,
		})
	}

	return OwnerAccounts200JSONResponse{
		AccountsSuccessJSONResponse: AccountsSuccessJSONResponse{
			Message: fmt.Sprintf("got %d accounts", len(acc)),
			Payload: acc,
		},
	}, err
}

// Get an account and their balances by pem certificate encoded in base64
// (GET /owner/accounts/me)

func (c Controller) OwnerMe(ctx context.Context, request OwnerMeRequestObject) (OwnerMeResponseObject, error) {
	var code string
	if request.Params.Code != nil {
		code = *request.Params.Code
	}

	var cred string
	var id string
	if request.Params.Cred == nil && request.Params.WalletId == nil {
		return OwnerMedefaultJSONResponse{
			Body: Error{
				Message: "can't get history",
				Payload: "one of the header 'wallet_id' or 'cred' is required",
			},
			StatusCode: 400,
		}, nil
	} else {
		if request.Params.Cred != nil {
			cred = *request.Params.Cred
		} else if request.Params.WalletId != nil {
			id = *request.Params.WalletId
		}
	}

	walId, balance, err := c.Service.GetBalanceByCredSigner(c.AppEnv.FSCName, c.AppEnv.COmpanyID, cred, id, code)
	if err != nil {
		return OwnerMedefaultJSONResponse{
			Body: Error{
				Message: "can't get account",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	amounts := []Amount{}
	for typ, val := range balance {
		amounts = append(amounts, Amount{
			Code:  typ,
			Value: val,
		})
	}

	return OwnerMe200JSONResponse{
		AccountSuccessJSONResponse: AccountSuccessJSONResponse{
			Message: fmt.Sprintf("got balances for %s", walId),
			Payload: Account{
				Id:      walId,
				Balance: amounts,
			},
		},
	}, nil
}

// Get an account and their balances
// (GET /owner/accounts/{id})
func (c Controller) OwnerAccount(ctx context.Context, request OwnerAccountRequestObject) (OwnerAccountResponseObject, error) {
	var code string
	if request.Params.Code != nil {
		code = *request.Params.Code
	}
	balance, err := c.Service.GetBalance(request.Id, code)
	if err != nil {
		return OwnerAccountdefaultJSONResponse{
			Body: Error{
				Message: "can't get account",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	amounts := []Amount{}
	for typ, val := range balance {
		amounts = append(amounts, Amount{
			Code:  typ,
			Value: val,
		})
	}
	return OwnerAccount200JSONResponse{
		AccountSuccessJSONResponse: AccountSuccessJSONResponse{
			Message: fmt.Sprintf("got balances for %s", request.Id),
			Payload: Account{
				Id:      request.Id,
				Balance: amounts,
			},
		},
	}, nil
}

// Get all transactions for an account
// (GET /owner/accounts/{id}/transactions)
func (c Controller) OwnerTransactions(ctx context.Context, request OwnerTransactionsRequestObject) (OwnerTransactionsResponseObject, error) {
	var history []service.TransactionHistoryItem
	var err error

	var param string
	var paramVal string
	if request.Params.TransactionCategory != nil {
		param = "transactionCategory"
		paramVal = *request.Params.TransactionCategory
	} else if request.Params.TransactionType != nil {
		param = "transactionType"
		paramVal = *request.Params.TransactionType
	} else {
		param = ""
		paramVal = ""
	}

	history, err = c.Service.GetHistory(request.Id, param, paramVal)
	if err != nil {
		return OwnerTransactionsdefaultJSONResponse{
			Body: Error{
				Message: "can't get history",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	pl := []TransactionRecord{}
	for _, tx := range history {
		TrxDetail := TransactionDetail{
			TransactionType:     tx.TransactionType,
			TransactionCategory: tx.TransactionCategory,
			Platform:            tx.Platform,
		}

		pl = append(pl, TransactionRecord{
			Amount: Amount{
				Code:  tx.TokenType,
				Value: tx.Amount,
			},
			Id:                tx.TxID,
			Recipient:         tx.Recipient,
			Sender:            tx.Sender,
			Status:            tx.Status,
			Timestamp:         tx.Timestamp,
			Message:           tx.Message,
			TransactionDetail: &TrxDetail,
		})
	}
	return OwnerTransactions200JSONResponse{
		TransactionsSuccessJSONResponse: TransactionsSuccessJSONResponse{
			Message: fmt.Sprintf("got %d transactions for %s", len(pl), request.Id),
			Payload: pl,
		},
	}, nil
}

func (c Controller) OwnerTransactionsMe(ctx context.Context, request OwnerTransactionsMeRequestObject) (OwnerTransactionsMeResponseObject, error) {
	var walId string
	var history []service.TransactionHistoryItem
	var err error

	var param string
	var paramVal string
	if request.Params.TransactionCategory != nil {
		param = "transactionCategory"
		paramVal = *request.Params.TransactionCategory
	} else if request.Params.TransactionType != nil {
		param = "transactionType"
		paramVal = *request.Params.TransactionType
	} else {
		param = ""
		paramVal = ""
	}

	var cred string
	var id string
	if request.Params.Cred == nil && request.Params.WalletId == nil {
		return OwnerTransactionsMedefaultJSONResponse{
			Body: Error{
				Message: "can't get history",
				Payload: "one of the header 'wallet_id' or 'cred' is required",
			},
			StatusCode: 400,
		}, nil
	} else {
		if request.Params.Cred != nil {
			cred = *request.Params.Cred
		} else if request.Params.WalletId != nil {
			id = *request.Params.WalletId
		}
	}

	walId, history, err = c.Service.GetHistoyByCred(c.AppEnv.FSCName, cred, id, param, paramVal)
	if err != nil {
		return OwnerTransactionsMedefaultJSONResponse{
			Body: Error{
				Message: "can't get history",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	pl := []TransactionRecord{}
	for _, tx := range history {
		TrxDetail := TransactionDetail{
			TransactionType:     tx.TransactionType,
			TransactionCategory: tx.TransactionCategory,
			Platform:            tx.Platform,
		}

		pl = append(pl, TransactionRecord{
			Amount: Amount{
				Code:  tx.TokenType,
				Value: tx.Amount,
			},
			Id:                tx.TxID,
			Recipient:         tx.Recipient,
			Sender:            tx.Sender,
			Status:            tx.Status,
			Timestamp:         tx.Timestamp,
			Message:           tx.Message,
			TransactionDetail: &TrxDetail,
		})
	}
	return OwnerTransactionsMe200JSONResponse{
		TransactionsSuccessJSONResponse: TransactionsSuccessJSONResponse{
			Message: fmt.Sprintf("got %d transactions for %s", len(pl), walId),
			Payload: pl,
		},
	}, nil
}

// Redeem (burn) tokens
// (POST /owner/accounts/{id}/redeem)
func (c Controller) Redeem(ctx context.Context, request RedeemRequestObject) (RedeemResponseObject, error) {
	code := request.Body.Amount.Code
	value := uint64(request.Body.Amount.Value)

	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	account := request.Id
	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	txID, err := c.Service.RedeemTokens(code, value, trxType, trxCat, platform, account, message)
	if err != nil {
		return RedeemdefaultJSONResponse{
			Body: Error{
				Message: "can't redeem tokens",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return Redeem200JSONResponse{
		RedeemSuccessJSONResponse: RedeemSuccessJSONResponse{
			Message: fmt.Sprintf("%s redeemed %d %s", account, value, code),
			Payload: txID,
		},
	}, nil
}

func (c Controller) RedeemMe(ctx context.Context, request RedeemMeRequestObject) (RedeemMeResponseObject, error) {
	code := request.Body.Amount.Code
	value := uint64(request.Body.Amount.Value)

	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	walId, txID, err := c.Service.RedeemTokenByCred(c.AppEnv.FSCName, request.Params.Cred, code, value, trxType, trxCat, platform, message)
	if err != nil {
		return RedeemMedefaultJSONResponse{
			Body: Error{
				Message: "can't redeem tokens",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return RedeemMe200JSONResponse{
		RedeemSuccessJSONResponse: RedeemSuccessJSONResponse{
			Message: fmt.Sprintf("%s redeemed %d %s", walId, value, code),
			Payload: txID,
		},
	}, nil
}
