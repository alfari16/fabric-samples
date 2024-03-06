package service

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/ttx"
	"github.com/pkg/errors"
)

type SignerConfig struct {
	Cred                      string `json:"Cred"`
	Sk                        string `json:"Sk"`
	EnrollmentID              string `json:"enrollment_id"`
	CredRevocationInformation string `json:"credential_revocation_information"`
	CurveID                   string `json:"curve_id"`
	RevocationHandle          string `json:"revocation_handle"`
}

type SignerData map[string]SignerConfig

func (s SignerData) isEmpty() bool {
	return len(s) == 0
}

var dataSignerById = make(SignerData)
var dataSignerByCred = make(SignerData)

func (s TokenService) getSignerDataByCred(fscName string, cred string) (*SignerConfig, error) {
	if dataSignerById.isEmpty() || dataSignerByCred.isEmpty() {
		cmd := fmt.Sprintf("ls /var/fsc/keys/%s/wallet", fscName)
		out, err := exec.Command("/bin/sh", "-c", cmd).Output()
		if err != nil {
			return nil, errors.Wrap(err, "Error listing keys")
		}

		listKeysWallet := strings.Split(string(out), "\n")

		for _, key := range listKeysWallet {
			walletName := strings.ReplaceAll(key, " ", "")
			if walletName == "" {
				continue
			}

			signerData, err := os.ReadFile(fmt.Sprintf("/var/fsc/keys/%s/wallet/%s/msp/user/SignerConfig", fscName, walletName))
			if err != nil {
				return nil, errors.Wrap(err, "Error reading signer config")
			}

			var signerConfig SignerConfig

			err = json.Unmarshal(signerData, &signerConfig)
			if err != nil {
				return nil, errors.Wrap(err, "Error unmarshal signer config")
			}

			dataSignerById[signerConfig.EnrollmentID] = signerConfig
			dataSignerByCred[signerConfig.Cred] = signerConfig
		}
	}

	signerConfig, ok := dataSignerByCred[cred]
	if !ok {
		return nil, nil
	}

	return &signerConfig, nil
}

func (s TokenService) getWallet(compId string, id string) (string, *token.OwnerWallet, error) {
	var walId = id

	w := ttx.GetWallet(s.FSC, id)
	if w == nil {
		walId = id + "@" + compId
		w = ttx.GetWallet(s.FSC, walId)
	}
	if w == nil {
		walId = id + "#" + compId
		w = ttx.GetWallet(s.FSC, walId)
	}

	if w == nil {
		return "", nil, errors.Errorf("wallet not found: %s", id)
	}

	return walId, w, nil
}
