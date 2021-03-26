// Hayek小工具
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"rsc.io/qr"

	"xcoin/HayekTool/pkg/config"
	"xcoin/HayekTool/pkg/mainpkg"
	"xcoin/HayekTool/pkg/util"
)

var (
	pkgVersion string = "(devel)"
	pkgCoinId  string = "hyk"
)

func main() {
	var app = cli.NewApp()

	app.Name = "HayekTool"
	app.Usage = "HayekTool"
	app.Version = pkgVersion

	app.UsageText = `HayekTool
   HayekTool [global options] command [command options] [arguments...]
   
   HayekTool help
   HayekTool -h`

	app.Authors = []*cli.Author{
		{
			Name:  "hyk",
			Email: "hayek@gmail.com",
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Value:   "config.json",
			Usage:   "HayekTool config file",
			EnvVars: []string{"HAYEK_TOOL_CONFIG"},
		},
		&cli.StringFlag{
			Name:  "coin-id",
			Value: "",
			Usage: "Set coin id",
		},
	}

	app.Before = func(c *cli.Context) error {
		if id := c.String("coin-id"); id != "" {
			pkgCoinId = id
		}

		if cfg := config.MustLoad(c.String("config")); cfg.DebugMode {
			log.SetFlags(log.LstdFlags | log.Llongfile)
		} else {
			log.SetFlags(log.LstdFlags)
		}

		return nil
	}

	app.Commands = []*cli.Command{
		{
			Name:  "gen-config",
			Usage: "gen config file",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "json format (default)",
				},
				&cli.BoolFlag{
					Name:  "toml",
					Usage: "toml format",
				},
			},

			Action: func(c *cli.Context) error {
				switch {
				case c.Bool("json"):
					fmt.Print(config.Default().JSONString())
				case c.Bool("toml"):
					fmt.Print(config.Default().TOMLString())
				default:
					fmt.Print(config.Default().JSONString())
				}
				return nil
			},
		},

		{
			Name:  "gen-address",
			Usage: "gen address",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "n",
					Usage: "set address number",
					Value: 1,
				},
				&cli.BoolFlag{
					Name:  "eip55",
					Usage: "use EIP55 format",
				},
				&cli.BoolFlag{
					Name:  "qrcode",
					Usage: "generate qrcode image",
					Value: true,
				},
				&cli.StringFlag{
					Name:  "outdir",
					Value: "zz_output_address",
					Usage: "Set output dir",
				},
			},

			Action: func(c *cli.Context) error {
				outbuf := new(bytes.Buffer)
				outdir := c.String("outdir")
				os.MkdirAll(outdir, 0777)
				for i := 1; i <= c.Int("n"); i++ {
					key, addr := util.GenAddress(c.Bool("eip55"))

					fmt.Printf("%03d %s %s\n", i, addr, key)
					fmt.Fprintf(outbuf, "%03d %s %s\n", i, addr, key)

					if outdir != "" {
						if c.Bool("qrcode") {
							s := strings.ToLower(addr)
							if m, err := qr.Encode(s, qr.H); err == nil {
								ioutil.WriteFile(
									filepath.Join(outdir, fmt.Sprintf("%03d_%s.png", i, s)),
									m.PNG(), 0644,
								)
							}
							if m, err := qr.Encode(key, qr.H); err == nil {
								ioutil.WriteFile(
									filepath.Join(outdir, fmt.Sprintf("%03d_%s_key.png", i, s)),
									m.PNG(), 0644,
								)
							}
						}
					}
				}
				ioutil.WriteFile(
					filepath.Join(outdir, "all.txt"),
					outbuf.Bytes(), 0644,
				)
				return nil
			},
		},

		{
			Name:  "get-work",
			Usage: "get work",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				return mainpkg.NewApp(cfg).CmdGetWork()
			},
		},

		{
			Name:  "get-pending-block",
			Usage: "get pending block",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				return mainpkg.NewApp(cfg).CmdGetPendingBlock()
			},
		},

		{
			Name:  "get-latest-block",
			Usage: "get latest block",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				return mainpkg.NewApp(cfg).CmdGetLatestBlock()
			},
		},

		{
			Name:  "get-balance",
			Usage: "get balance",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.StringFlag{
					Name:  "address",
					Usage: "set address",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				address := cfg.UserAddress
				if s := c.String("address"); s != "" {
					address = s
				}

				return mainpkg.NewApp(cfg).CmdGetBalance(address)
			},
		},

		{
			Name:  "get-peer-count",
			Usage: "get peer count",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				return mainpkg.NewApp(cfg).CmdGetPeerCount()
			},
		},
		{
			Name:  "get-net-version",
			Usage: "get net version",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				return mainpkg.NewApp(cfg).CmdNetVersion()
			},
		},

		{
			Name:  "get-tx",
			Usage: "get tx receipt",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.StringFlag{
					Name:  "hash",
					Usage: "tx hash",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				hash := c.String("hash")
				if hash == "" {
					fmt.Println("no tx hash")
					os.Exit(1)
				}

				return mainpkg.NewApp(cfg).CmdGetTxReceipt(hash)
			},
		},

		{
			Name:  "get-block-by-hash",
			Usage: "get block by hash",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.StringFlag{
					Name:  "hash",
					Usage: "block hash",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				hash := c.String("hash")
				if hash == "" {
					fmt.Println("no block hash")
					os.Exit(1)
				}

				return mainpkg.NewApp(cfg).CmdGetBlockByHash(hash)
			},
		},

		{
			Name:  "get-block-by-height",
			Usage: "get block by height",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.IntFlag{
					Name:  "height",
					Usage: "block height",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				height := c.Int("height")
				if height <= 0 {
					fmt.Println("no block height")
					os.Exit(1)
				}

				return mainpkg.NewApp(cfg).CmdGetBlockByHeight(int64(height))
			},
		},

		{
			Name:  "send-tx",
			Usage: "send tx",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.StringFlag{
					Name:  "to",
					Usage: "set send to address",
				},
				&cli.IntFlag{
					Name:  "value",
					Usage: "set send tx value",
				},
				&cli.IntFlag{
					Name:  "gas-limit",
					Usage: "set gas limit",
					Value: mainpkg.DefaultGasLimit,
				},
				&cli.IntFlag{
					Name:  "gas-price",
					Usage: "set gas price",
					Value: mainpkg.DefaultGasPrice,
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				to := c.String("to")
				if to == "" {
					fmt.Println("missing to address")
					os.Exit(1)
				}

				return mainpkg.NewApp(cfg).CmdSendTx(to,
					c.Int64("value"),
					c.Int64("gas-limit"),
					c.Int64("gas-price"),
				)
			},
		},

		{
			Hidden: true, // 内部功能

			Name:  "send-payouts",
			Usage: "send payouts",

			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "host",
					Usage: "set host url",
				},
				&cli.StringFlag{
					Name:  "payouts-file",
					Usage: "set payouts file",
					Value: "payouts-file.json",
				},
			},

			Action: func(c *cli.Context) error {
				cfg := config.MustLoad(c.String("config"))
				if s := c.String("host"); s != "" {
					cfg.Host = s
				}

				mainpkg.NewApp(cfg).CmdRunPayoutsService(
					c.String("payouts-file"),
				)
				return nil
			},
		},
	}

	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fmt.Fprintf(ctx.App.Writer, "not found '%v'!\n", command)
	}

	app.Run(os.Args)
}
