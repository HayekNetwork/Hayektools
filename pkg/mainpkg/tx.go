package mainpkg

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"xcoin/HayekTool/pkg/rpc"
	"xcoin/HayekTool/pkg/util"
)

const (
	DefaultGasLimit = 21000
	DefaultGasPrice = 100000000000
)

func (p *App) CmdSendTx(to string, value, gasLimit, gasPrice int64) error {
	if value == 0 {
		return fmt.Errorf("invalue value")
	}
	if gasLimit == 0 {
		gasLimit = DefaultGasLimit
	}
	if gasPrice == 0 {
		gasPrice = DefaultGasPrice
	}

	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		return err
	}

	valueWei := new(big.Int).Mul(big.NewInt(value), util.Ether)

	to = p.cfg.GetAddress(to)

	txHash, err := p.sendRawTx(c, to, valueWei, uint64(gasLimit), big.NewInt(gasPrice))
	if err != nil {
		return err
	}

	fmt.Println("txHash:", txHash)
	return nil
}

func (p *App) sendRawTx(
	client *rpc.RPCClient, to string, value *big.Int,
	gasLimit uint64, gasPrice *big.Int,
) (
	txHash string, err error,
) {
	var (
		privateKey, _ = crypto.HexToECDSA(p.cfg.UserKey)
		toAddress     = common.HexToAddress(to)
	)

	if p.cfg.DebugMode {
		s, _ := json.MarshalIndent(p.cfg, "", "\t")
		fmt.Printf("App.sendRawTx: p.cfg = %s\n", s)
	}

	if privateKey == nil {
		if p.cfg.DebugMode {
			log.Println("err")
		}
		return "", fmt.Errorf("invalid UserKey")
	}

	nonce, err := client.GetTransactionCount(p.cfg.UserAddress, "pending")
	if err != nil {
		if p.cfg.DebugMode {
			log.Println("err")
		}
		return "", err
	}
	if p.cfg.DebugMode {
		log.Println("nonce:", nonce)
	}

	// nonce = 1
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	if p.cfg.DebugMode {
		s, _ := json.MarshalIndent(tx, "", "\t")
		log.Printf("App.sendRawTx: tx = %s\n", s)

		addr, err2 := types.NewEIP155Signer(rpc.ChainID).Sender(tx)
		log.Printf("App.sendRawTx111: %v, %x, %v", rpc.ChainID, addr, err2)
		if err2 != nil {
			if p.cfg.DebugMode {
				log.Println(err2)
			}
			return "", err2
		}
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(rpc.ChainID), privateKey)
	if err != nil {
		if p.cfg.DebugMode {
			log.Println(err)
		}
		return "", err
	}
	if p.cfg.DebugMode {
		s, _ := json.MarshalIndent(signedTx, "", "\t")
		log.Println("signedTx:", string(s))
	}

	data, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		if p.cfg.DebugMode {
			log.Println(err)
		}
		return "", err
	}
	if p.cfg.DebugMode {
		log.Println("EncodeToBytes:data:", string(data))
	}

	txHash, err = client.SendRawTransaction(hexutil.Encode(data))
	if err != nil {
		if p.cfg.DebugMode {
			log.Println(err)
		}
		return "", err
	}

	if s := signedTx.Hash().Hex(); txHash != s {
		if p.cfg.DebugMode {
			log.Println("err")
		}
		return "", fmt.Errorf("invalid tx hash: expect = %s, got = %s", s, txHash)
	}

	return txHash, nil
}
