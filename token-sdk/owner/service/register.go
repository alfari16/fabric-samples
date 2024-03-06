package service

import (
	"fmt"

	viewregistry "github.com/hyperledger-labs/fabric-smart-client/platform/view"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/pkg/errors"
)

func (s TokenService) RegisterOwnerWallet(fscName string, id string, username string) (string, error) {
	logger.Infof("register user wallet %s to %s", username, fscName)

	path := fmt.Sprintf("/var/fsc/keys/%s/wallet/%s/msp", fscName, username)

	// register wallet
	_, err := viewregistry.GetManager(s.FSC).InitiateView(&RegisterOwnerWalletView{
		RegisterOwner: &RegisterOwner{
			ID:   id,
			Path: path,
			TMSID: token.TMSID{
				Network:   "mynetwork",
				Channel:   "mychannel",
				Namespace: "tokenchaincode",
			},
		},
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to register wallet")
	}

	return path, nil
}

type RegisterOwner struct {
	ID    string
	Path  string // path msp
	TMSID token.TMSID
}

type RegisterOwnerWalletView struct {
	*RegisterOwner
}

func (r *RegisterOwnerWalletView) Call(context view.Context) (interface{}, error) {
	tms := token.GetManagementService(context, token.WithTMSID(r.TMSID))
	if tms == nil {
		return nil, errors.Errorf("tms not found")
	}

	err := tms.WalletManager().RegisterOwnerWallet(r.ID, r.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to register wallet")
	}

	// getwallet
	return nil, nil
}
