package mainpkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"

	"xcoin/HayekTool/pkg/clockwork"
	"xcoin/HayekTool/pkg/rpc"
	"xcoin/HayekTool/pkg/util"
)

// 用于定时给多个客户按比例分红文件
type PayoutsFile struct {
	Threshold     int64        // CoinBase 最小余额
	FeePercentage float64      // 消费(保留的比例)
	EveryDatAt    []string     // 每天定时触发的时间, 时间格式 hour:min, 比如 18:30 或 10:30 等
	GasLimit      int64        // Gas限制
	GasPrice      int64        // Gas价格
	Payouts       []PayoutElem // 支付列表
}

// 每个支付的地址和比例
type PayoutElem struct {
	Name            string  // 客户名字
	Address         string  // 客户地址
	ValuePercentage float64 // 支付比例(0.01～1.0)
}

func (p *App) CmdRunPayoutsService(payoutsFile string) {
	payoutsInfo, err := p.loadPayoutsFile(payoutsFile)
	if err != nil {
		p.genPayoutsFileTemplate(strings.TrimSuffix(payoutsFile, ".json") + ".example.json")
		log.Fatal(err)
	}

	if err := p.checkPayoutsFile(payoutsInfo); err != nil {
		log.Fatal(err)
	}

	for _, time := range payoutsInfo.EveryDatAt {
		clockwork.Every(1).Day().At(time).Do(func() {
			if err := p.doPayoutsTask(payoutsInfo); err != nil {
				log.Println(err)
			}
		})
	}

	// 启动定时任务
	<-clockwork.Start()
}

// 执行一次支付任务
func (p *App) doPayoutsTask(info *PayoutsFile) error {
	c, err := rpc.NewRPCClient("HayekTool", p.cfg.Host, time.Second*3)
	if err != nil {
		return err
	}

	amountInWei, err := c.GetBalance(p.cfg.UserAddress)
	if err != nil {
		return nil
	}

	if info.Threshold > 0 {
		vThresholdWei := new(big.Int).Mul(big.NewInt(info.Threshold), util.Ether)
		if amountInWei.Cmp(vThresholdWei) < 0 {
			return fmt.Errorf("balance limit: threshold = %v", info.Threshold)
		}
	}

	var (
		gasLimit = info.GasLimit
		gasPrice = info.GasPrice
	)
	if gasLimit == 0 {
		gasLimit = DefaultGasLimit
	}
	if gasPrice == 0 {
		gasPrice = DefaultGasPrice
	}

	for _, to := range info.Payouts {
		xRate := to.ValuePercentage

		xReward := new(big.Float).Mul(new(big.Float).SetInt(amountInWei), big.NewFloat(xRate))

		xRewardWei := xReward.Mul(xReward, new(big.Float).SetInt(util.Ether))

		valueWei, _ := xRewardWei.Int(nil)

		txHash, err := p.sendRawTx(c, to.Address, valueWei, uint64(gasLimit), big.NewInt(gasPrice))
		if err != nil {
			return err
		}

		fmt.Println("txHash:", txHash)
	}

	return nil
}

func (p *App) checkPayoutsFile(info *PayoutsFile) error {
	if len(info.EveryDatAt) == 0 {
		return fmt.Errorf("empty EveryDatAt")
	}
	if len(info.Payouts) == 0 {
		return fmt.Errorf("empty Payouts")
	}

	// 验证是否大于 100%
	var totalValuePercentage = info.FeePercentage
	for _, v := range info.Payouts {
		totalValuePercentage += v.ValuePercentage
	}
	if totalValuePercentage > 1 {
		return fmt.Errorf("ValuePercentage overflow(%v)", totalValuePercentage)
	}

	return nil
}

func (p *App) loadPayoutsFile(payoutsFile string) (*PayoutsFile, error) {
	data, err := ioutil.ReadFile(payoutsFile)
	if err != nil {
		return nil, err
	}

	var info PayoutsFile
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (p *App) genPayoutsFileTemplate(payoutsFile string) error {
	x := &PayoutsFile{
		Threshold:     1,
		FeePercentage: 0.10,
		EveryDatAt:    []string{"10:30", "18:30"},
		GasLimit:      0,
		GasPrice:      0,
		Payouts: []PayoutElem{
			{
				Name:            "user0",
				Address:         "0x3eb41fc94f240242c9bbb8bf46b9feb356fd09e2",
				ValuePercentage: 0.1,
			},
		},
	}

	data, _ := json.MarshalIndent(x, "", "\t")
	return ioutil.WriteFile(payoutsFile, data, 0644)
}
