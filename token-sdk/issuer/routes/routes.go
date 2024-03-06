/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package routes

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-samples/token-sdk/issuer/service"
)

type Controller struct {
	Service service.TokenService
}

// Issue tokens to an account
// (POST /issue)
func (c Controller) Issue(ctx context.Context, request IssueRequestObject) (IssueResponseObject, error) {
	code := request.Body.Amount.Code
	value := uint64(request.Body.Amount.Value)
	recipient := request.Body.Counterparty.Account
	recipientNode := request.Body.Counterparty.Node
	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	txID, qty, err := c.Service.Issue(code, value, trxType, trxCat, platform, recipient, recipientNode, message)
	if err != nil {
		return IssuedefaultJSONResponse{
			Body: Error{
				Message: "can't issue tokens",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return Issue200JSONResponse{
		IssueSuccessJSONResponse: IssueSuccessJSONResponse{
			Message: fmt.Sprintf("issued %d %s to %s on %s", qty, code, recipient, recipientNode),
			Payload: txID,
		},
	}, nil
}

func (c Controller) Harvest(ctx context.Context, request HarvestRequestObject) (HarvestResponseObject, error) {
	code := "IDR"
	value := uint64(request.Body.Amount.Value)
	recipient := request.Body.Counterparty.Account
	recipientNode := request.Body.Counterparty.Node
	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	txID, qty, err := c.Service.Issue(code, value, trxType, trxCat, platform, recipient, recipientNode, message)
	if err != nil {
		return HarvestdefaultJSONResponse{
			Body: Error{
				Message: "can't issue tokens",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return Harvest200JSONResponse{
		IssueSuccessJSONResponse: IssueSuccessJSONResponse{
			Message: fmt.Sprintf("issued %d %s to %s on %s", qty, code, recipient, recipientNode),
			Payload: txID,
		},
	}, nil
}

func (c Controller) Kabayan(ctx context.Context, request KabayanRequestObject) (KabayanResponseObject, error) {
	code := "KBY"
	value := uint64(request.Body.Amount.Value)
	recipient := request.Body.Counterparty.Account
	recipientNode := request.Body.Counterparty.Node
	trxCat := request.Body.TransactionDetail.TransactionCategory
	trxType := request.Body.TransactionDetail.TransactionType
	platform := request.Body.TransactionDetail.Platform

	var message string
	if request.Body.Message != nil {
		message = *request.Body.Message
	}

	txID, qty, err := c.Service.Issue(code, value, trxType, trxCat, platform, recipient, recipientNode, message)
	if err != nil {
		return KabayandefaultJSONResponse{
			Body: Error{
				Message: "can't issue tokens",
				Payload: err.Error(),
			},
			StatusCode: 500,
		}, nil
	}

	return Kabayan200JSONResponse{
		IssueSuccessJSONResponse: IssueSuccessJSONResponse{
			Message: fmt.Sprintf("issued %d %s to %s on %s", qty, code, recipient, recipientNode),
			Payload: txID,
		},
	}, nil
}
