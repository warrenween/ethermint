package app

import (
	"encoding/json"
<<<<<<< HEAD
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

=======

	eRPC "github.com/ethereum/go-ethereum/rpc"

	abci "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"

>>>>>>> c074676... Restructure app.go
	"github.com/tendermint/ethermint/ethereum"
	emtTypes "github.com/tendermint/ethermint/types"

	abciTypes "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"
)

// EthermintApplication implements an ABCI application.
// It holds a reference to an ethereum backend, which can be implemented through
// various means, as long as it satisfies the ethereum.Backend interface.
// Furthermore, it also holds the strategy for this app. A strategy describes
// how to distribute rewards, such as block rewards and transaction fees, as
// well as to whom they should be given. A strategy also deals with validator
// set changes.
// #stable - 0.4.0
type EthermintApplication struct {
	// backend handles the ethereum state machine
	// and wrangles other services started by an ethereum node (eg. tx pool)
	// This is the running ethereum node.
	backend *ethereum.Backend // backend ethereum struct

	// an ethereum rpc client we can forward queries to
	// this client talks to the backend struct above
	eRPC *eRPC.Client

	// strategy for validator compensation
	strategy *emtTypes.Strategy

	logger tmLog.Logger
}

// NewEthermintApplication creates a fully initialised instance of EthermintApplication
// #stable - 0.4.0
func NewEthermintApplication(backend *ethereum.Backend,
	eRPC *eRPC.Client, strategy *emtTypes.Strategy,
	logger tmLog.Logger) (*EthermintApplication, error) {
	app := &EthermintApplication{
		backend:  backend,
		eRPC:     eRPC,
		strategy: strategy,
		logger:   logger,
	}

	if err := app.backend.ResetWork(app.Receiver()); err != nil {
		return nil, err
	}

	return app, nil
}

// -------------------------
// Info/Query Connection

var bigZero = big.NewInt(0)

// Info returns information about the last height and app_hash to the tendermint engine
// #stable - 0.4.0
<<<<<<< HEAD
func (app *EthermintApplication) Info() abciTypes.ResponseInfo {
	blockchain := app.backend.Ethereum().BlockChain()
	currentBlock := blockchain.CurrentBlock()
	height := currentBlock.Number()
	hash := currentBlock.Hash()

	app.logger.Debug("Info", "height", height) // nolint: errcheck

	// This check determines whether it is the first time ethermint gets started.
	// If it is the first time, then we have to respond with an empty hash, since
	// that is what tendermint expects.
	if height.Cmp(bigZero) == 0 {
		return abciTypes.ResponseInfo{
			Data:             "ABCIEthereum",
			LastBlockHeight:  height.Uint64(),
			LastBlockAppHash: []byte{},
		}
	}

	return abciTypes.ResponseInfo{
		Data:             "ABCIEthereum",
		LastBlockHeight:  height.Uint64(),
		LastBlockAppHash: hash[:],
	}
=======
func (a *EthermintApplication) Info() abci.ResponseInfo {
	return a.backend.Info()
>>>>>>> c074676... Restructure app.go
}

// SetOption sets a configuration option
// #stable - 0.4.0
func (a *EthermintApplication) SetOption(key string, value string) string {
	return a.backend.SetOption(key, value)
}

// Query queries the state of the EthermintApplication
// #stable - 0.4.0
func (a *EthermintApplication) Query(query abci.RequestQuery) abci.ResponseQuery {
	var in jsonRequest
	if err := json.Unmarshal(query.Data, &in); err != nil {
		return abci.ResponseQuery{Code: abci.ErrEncodingError.Code, Log: err.Error()}
	}
	var result interface{}
	if err := a.eRPC.Call(&result, in.Method, in.Params...); err != nil {
		return abci.ResponseQuery{Code: abci.ErrInternalError.Code, Log: err.Error()}
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return abci.ResponseQuery{Code: abci.ErrInternalError.Code, Log: err.Error()}
	}
	return abci.ResponseQuery{Code: abci.OK.Code, Value: bytes}
}

// ---------------------------
// Mempool Connection

// CheckTx checks a transaction is valid but does not mutate the state
// #stable - 0.4.0
func (a *EthermintApplication) CheckTx(txBytes []byte) abci.Result {
	tx, err := decodeTx(txBytes)
	a.logger.Debug("CheckTx: Received valid transaction", "tx", tx) // nolint: errcheck
	if err != nil {
		a.logger.Debug("CheckTx: Received invalid transaction", "tx", tx) // nolint: errcheck
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

<<<<<<< HEAD
	return app.validateTx(tx)
=======
	return a.backend.CheckTx(tx)
}

// ---------------------
// Consensus Connection

// InitChain initializes the validator set
// #stable - 0.4.0
func (a *EthermintApplication) InitChain(validators []*abci.Validator) {
	a.logger.Debug("InitChain") // nolint: errcheck
	a.SetValidators(validators)
}

// BeginBlock starts a new Ethereum block
// #stable - 0.4.0
func (a *EthermintApplication) BeginBlock(hash []byte, tmHeader *abci.Header) {
	a.logger.Debug("BeginBlock") // nolint: errcheck

	// update the eth header with the tendermint header
	a.backend.UpdateHeaderWithTimeInfo(tmHeader)
>>>>>>> c074676... Restructure app.go
}

// DeliverTx executes a transaction against the latest state
// #stable - 0.4.0
func (a *EthermintApplication) DeliverTx(txBytes []byte) abci.Result {
	tx, err := decodeTx(txBytes)
	if err != nil {
		a.logger.Debug("DelivexTx: Received invalid transaction", "tx", tx, "err", err) // nolint: errcheck
		return abci.ErrEncodingError.AppendLog(err.Error())
	}
	a.logger.Debug("DeliverTx: Received valid transaction", "tx", tx) // nolint: errcheck

	err = a.backend.DeliverTx(tx)
	if err != nil {
		a.logger.Error("DeliverTx: Error delivering tx to ethereum backend", "tx", tx, "err", err) // nolint: errcheck
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	a.CollectTx(tx)

	return abci.OK
}

// EndBlock accumulates rewards for the validators and updates them
// #stable - 0.4.0
func (a *EthermintApplication) EndBlock(height uint64) abci.ResponseEndBlock {
	a.logger.Debug("EndBlock", "height", height) // nolint: errcheck
	a.backend.AccumulateRewards(a.strategy)
	return a.GetUpdatedValidators()
}

// Commit commits the block and returns a hash of the current state
// #stable - 0.4.0
func (a *EthermintApplication) Commit() abci.Result {
	a.logger.Debug("Commit") // nolint: errcheck
	blockHash, err := a.backend.Commit(a.Receiver())
	if err != nil {
		a.logger.Error("Error getting latest ethereum state", "err", err) // nolint: errcheck
		return abci.ErrInternalError.AppendLog(err.Error())
	}
	return abci.NewResultOK(blockHash[:], "")
}

//-------------------------------------------------------

// validateTx checks the validity of a tx against the blockchain's current state.
// it duplicates the logic in ethereum's tx_pool
func (app *EthermintApplication) validateTx(tx *ethTypes.Transaction) abciTypes.Result {
	currentState, err := app.currentState()
	if err != nil {
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return abciTypes.ErrBaseUnknownAddress.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Check for nonce errors
	currentNonce := currentState.GetNonce(from)
	if currentNonce > tx.Nonce() {
		return abciTypes.ErrBadNonce.
			AppendLog(fmt.Sprintf("Got: %d, Current: %d", tx.Nonce(), currentNonce))
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.AppendLog(core.ErrGasLimitReached.Error())
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Cmp(common.Big0) < 0 {
		return abciTypes.ErrBaseInvalidInput.
			SetLog(core.ErrNegativeValue.Error())
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	currentBalance := currentState.GetBalance(from)
	if currentBalance.Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(fmt.Sprintf("Current balance: %s, tx cost: %s", currentBalance, tx.Cost()))

	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}

	return abciTypes.OK
}
