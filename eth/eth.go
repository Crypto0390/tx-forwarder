package eth

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	samplecontract "github.com/iden3/gas-station/eth/contract"
	log "github.com/sirupsen/logrus"
)

type EthService struct {
	ks              *keystore.KeyStore
	acc             *accounts.Account
	client          *ethclient.Client
	SampleContract  *samplecontract.SampleContract
	contractAddress common.Address
	KeyStore        struct {
		Path     string
		Password string
	}
}

func (ethSrv *EthService) Account() *accounts.Account {
	return ethSrv.acc
}

func NewEthService(url string, ks *keystore.KeyStore, acc *accounts.Account, keystorePath, password string) *EthService {
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Error("Can not open connection to web3 (config.Web3.Url: " + url + ")\n" + err.Error() + "\n")
		os.Exit(0)
	}
	log.Info("Connection to web3 server opened")

	service := &EthService{
		ks:     ks,
		acc:    acc,
		client: client,
		KeyStore: struct {
			Path     string
			Password string
		}{
			Path:     keystorePath,
			Password: password,
		},
	}

	return service
}

func (ethSrv *EthService) GetBalance(address common.Address) (*big.Float, error) {
	balance, err := ethSrv.client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return nil, err
	}
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethBalance := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	return ethBalance, nil
}

func (ethSrv *EthService) DeployContract() error {
	auth, err := ethSrv.GetAuth()
	if err != nil {
		return err
	}

	address, tx, _, err := samplecontract.DeploySampleContract(auth, ethSrv.client)
	if err != nil {
		return err
	}
	ethSrv.contractAddress = address

	log.Info("sample contract deployed at address: " + address.Hex())
	fmt.Println("deployment transaction: " + tx.Hash().Hex())
	return nil
}

func (ethSrv *EthService) LoadContract(contractAddr common.Address) {
	instance, err := samplecontract.NewSampleContract(contractAddr, ethSrv.client)
	if err != nil {
		log.Error(err.Error())
	}
	ethSrv.contractAddress = contractAddr
	ethSrv.SampleContract = instance
}

func (ethSrv *EthService) GetAuth() (*bind.TransactOpts, error) {
	file, err := os.Open(ethSrv.KeyStore.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewTransactor(strings.NewReader(string(b)), ethSrv.KeyStore.Password)
	return auth, err
}
