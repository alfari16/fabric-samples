package service

import (
	"fmt"
	"os"
	"os/exec"

	viewregistry "github.com/hyperledger-labs/fabric-smart-client/platform/view"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/ttx"
	"github.com/pkg/errors"
)

func (s TokenService) RegisterOwnerWallet(fscName string, id string, username string) (string, error) {
	path := fmt.Sprintf("/var/fsc/keys/%s/wallet/%s/msp", fscName, id)

	// checking before process
	// check wallet is exist in path or not, and check wallet is exist on owner or not
	_, err := os.Stat(path)

	// if w != nil && !os.IsNotExist(err) {
	// 	return fmt.Sprintf("wallet already exist: %s", id), nil
	// }
	host := os.Getenv("FABRIC_CA_HOST")

	if os.IsNotExist(err) {
		args := []string{
			"register",
			"-u",
			fmt.Sprintf("http://%s", host),
			"--id.attrs",
			fmt.Sprintf("full_name=%s", username),
			"--id.name",
			id,
			"--id.secret",
			"password",
			"--id.type",
			"client",
			"--enrollment.type",
			"idemix",
			"--idemix.curve",
			"gurvy.Bn254",
		}
		o, err := exec.Command("/tmp/fabric-ca-client", args...).CombinedOutput()
		if err != nil {
			logger.Errorf("error when register with fabric-ca-client: %s", err.Error())
			return "", err
		}
		logger.Infof("successfully register using fabric-ca-client: %s", o)

		args = []string{
			"enroll",
			"-u",
			fmt.Sprintf("http://%s:password@%s", id, host),
			"-M",
			path,
			"--enrollment.type",
			"idemix",
			"--idemix.curve",
			"gurvy.Bn254",
		}
		o, err = exec.Command("/tmp/fabric-ca-client", args...).CombinedOutput()
		if err != nil {
			logger.Errorf("error when enroll with fabric-ca-client: %s", err.Error())
			return "", err
		}
		logger.Infof("successfully enroll using fabric-ca-client: %s-%s", path, o)
	}

	logger.Infof("register user wallet %s to %s, path: %s", username, fscName, path)

	// ketika kita register ke chaincode
	w := ttx.GetWallet(s.FSC, id)

	if w != nil {
		return fmt.Sprintf("wallet already exist: %s", id), nil
	}

	// register wallet
	_, err = viewregistry.GetManager(s.FSC).InitiateView(&RegisterOwnerWalletView{
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
