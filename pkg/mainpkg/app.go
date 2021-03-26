package mainpkg

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"xcoin/HayekTool/pkg/config"
	"xcoin/HayekTool/pkg/rpc"
	"xcoin/HayekTool/pkg/util"
)

type App struct {
	cfg *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (p *App) CmdGetWork() error {

	var lastErr error
	var lastWork []string

	for {
		client, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
		if err != nil && lastErr == nil {
			log.Fatal(err)
		}
		if lastErr != nil && err == nil {
			log.Printf("connect %s ok\n", p.cfg.Host)
		}

		lastErr = err

		for {
			work, err := client.GetWork()
			if err != nil {
				log.Printf("GetWork: err = %v", err)
				break
			}
			if len(work) != 6 {
				log.Printf("GetWork: invalid work, len != 6")
				break
			}

			if strings.Join(work, ",") != strings.Join(lastWork, ",") {
                lastWork = work
                type GetWorkReply struct {
                    Header    string // reply[0]
                    Seed      string // reply[1]
					Target    string // reply[2]
					Height    string // reply[3]
					StateRoot string // reply[4]
					Timestamp string // reply[5]

					XXXHeight    int
					XXXTimestamp time.Time
				}

				data := GetWorkReply{
					Header:    work[0],
					Seed:      work[1],
					Target:    work[2],
					Height:    work[3],
					StateRoot: work[4],
					Timestamp: work[5],
				}

				height, _ := strconv.ParseUint(strings.Replace(work[3], "0x", "", -1), 16, 64)
				timestamp, _ := strconv.ParseUint(strings.Replace(work[5], "0x", "", -1), 16, 64)

				data.XXXHeight = int(height)
				data.XXXTimestamp = time.Unix(int64(timestamp), 0)

				s, _ := json.MarshalIndent(data, "", "\t")
				fmt.Println(string(s))

				time.Sleep(time.Second * 3)

			} else {
				time.Sleep(time.Second)
			}
		}
	}
}

func (p *App) CmdGetPendingBlock() error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetPendingBlock(true)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetLatestBlock() error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetLatestBlock(true)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetBlockByHeight(height int64) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetBlockByHeight(height, true)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetBlockByHash(hash string) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetBlockByHash(hash, true)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetUncleByBlockNumberAndIndex(height int64, index int) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetUncleByBlockNumberAndIndex(height, index)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetTxReceipt(hash string) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.GetTxReceipt(hash)
	if err != nil {
		log.Fatal(err)
	}

	s, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(s))
	return nil
}

func (p *App) CmdGetBalance(idOrAddress string) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	address := p.cfg.GetAddress(idOrAddress)
	amountInWei, err := c.GetBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"%d(wei), %.6f(Gwei), %.6f(Ether)\n",
		amountInWei,
		new(big.Float).Quo(new(big.Float).SetInt(amountInWei), new(big.Float).SetInt(util.Shannon)),
		new(big.Float).Quo(new(big.Float).SetInt(amountInWei), new(big.Float).SetInt(util.Ether)),
	)

	return nil
}

func (p *App) CmdGetPeerCount() error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	v, err := c.GetPeerCount()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v)
	return nil
}

func (p *App) CmdNetVersion() error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}

	v, err := c.NetVersion()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(v)
	return nil
}

/*
func (r *RPCClient) SubmitBlock(params []string) (bool, error) {
func (r *RPCClient) Sign(from string, s string) (string, error) {
func (r *RPCClient) SendTransaction(from, to, gas, gasPrice, value string, autoGas bool) (string, error) {
*/
