# Hayek小工具

Hayek小工具用于钱包地址批量生成/转账/查余额/查区块信息等常用等基本功能.

## 构建程序

- 安装Go1.15+
- 安装GCC环境
- `go build`生成`HayekTool`可执行程序

## 基本用法

```
$ HayekTool -h
NAME:
   HayekTool - HayekTool

USAGE:
   HayekTool
   HayekTool [global options] command [command options] [arguments...]
   
   HayekTool help
   HayekTool -h

VERSION:
   (devel)

COMMANDS:
   gen-config           gen config file
   gen-address          gen address
   get-work             get work
   get-pending-block    get pending block
   get-latest-block     get latest block
   get-balance          get balance
   get-peer-count       get peer count
   get-tx               get tx receipt
   get-block-by-hash    get block by hash
   get-block-by-height  get block by height
   ...
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value   HayekTool config file (default: "config.json") [$HAYEK_TOOL_CONFIG]
   --coin-id value  Set coin id
   --help, -h       show help (default: false)
   --version, -v    print the version (default: false)
$
```

查看帮助:

```
$ HayekTool gen-config -h
NAME:
   HayekTool gen-config - gen config file

USAGE:
   HayekTool gen-config [command options] [arguments...]

OPTIONS:
   --json      json format (default) (default: false)
   --toml      toml format (default: false)
   --help, -h  show help (default: false)
```



生成JSON格式配置文件模板:

```json
$ HayekTool gen-config
{
        "DebugMode": false,
        "Host": "http://127.0.0.1:28585",
        "UserName": "Hayek",
        "UserKey": "",
        "UserAddress": "0x3eb41fc94f240242c9bbb8bf46b9feb356fd09e2",
        "XUserAddressBook": null
}
```

生成TOML格式配置文件:

```toml
$ HayekTool gen-config -toml
# HayekTool

DebugMode = false
Host = "http://127.0.0.1:28585"
UserName = "Hayek"
UserKey = ""
UserAddress = "0x3eb41fc94f240242c9bbb8bf46b9feb356fd09e2"
```

## 生成钱包地址

查看帮助:

```
$ HayekTool gen-address -h
NAME:
   HayekTool gen-address - gen address

USAGE:
   HayekTool gen-address [command options] [arguments...]

OPTIONS:
   -n value        set address number (default: 1)
   --eip55         use EIP55 format (default: false)
   --qrcode        generate qrcode image (default: true)
   --outdir value  Set output dir (default: "zz_output_address")
   --help, -h      show help (default: false)
```

其中`-n`表示要生成的钱包地址数码, 比如`-n=3`表示生成3个地址. `--eip55`表示EIP55规范生成校验码信息,
该规范地址中混合大小写字母进行校验. `-qrcode`表示生成二维码信息. `-outdir`表示生成的钱包地址保持的目录,
默认是`zz_output_address`目录.

生成一个钱包地址:

```
$ HayekTool gen-address
001 0x67eb7de0deca3fcdff57e0a154e5d2d70c047af8 95dbb6daf28de17f148748a7ff2826c9e792680575321ec857390811444c1727
```

第一列是钱包地址的序号, 第二列是钱包地址, 第三列是钱包地址对应的私钥.

批量生成多个地址:

```
$ HayekTool gen-address -n=3
001 0x983c7e8cd28fccafe313e4ceca8d5a617c9374a9 716874e46f60746d6a57d48e837bea3d7dca4c991a9d4859d1877588bb5f85c0
002 0x4e8f834e07a795c408f78a053ddb981f3dd14179 1a6dd52f293710fd50a2ef10bef6fe416a7bc9166889ba1235c250b682d0a8fc
003 0xae6393049ad5e19523e0d1245517b8cf7a03f21b 0f4b4bcbf72a02abeff0b291df6c07ae5ad2af284444277bb16974b315649be5
```

## 从主链获取任务

查看帮助:

```
$ HayekTool get-work -h
NAME:
   HayekTool get-work - get work

USAGE:
   HayekTool get-work [command options] [arguments...]

OPTIONS:
   --host value  set host url
   --help, -h    show help (default: false)
```
## 查询余额

查看帮助:

```
$ HayekTool get-balance -h
NAME:
   HayekTool get-balance - get balance

USAGE:
   HayekTool get-balance [command options] [arguments...]

OPTIONS:
   --host value     set host url
   --address value  set address
   --help, -h       show help (default: false)
```

获取当前账号余额:

```
$ HayekTool get-balance
4823991599994552600000(wei), 4823991599994.552600(Gwei), 4823.991600(Ether)
```

分别以不同的单位显示账户的余额.


## 获取peer节点数目

查看帮助:

```
$ HayekTool get-peer-count -h
NAME:
   HayekTool get-peer-count - get peer count

USAGE:
   HayekTool get-peer-count [command options] [arguments...]

OPTIONS:
   --host value  set host url
   --help, -h    show help (default: false)
```

获取peer节点数目:

```
$ HayekTool get-peer-count
3
```

## 获取 net-version

```
$ HayekTool get-net-version
```

## 获取最新的区块


查看帮助:

```
$ HayekTool get-latest-block -h
NAME:
   HayekTool get-latest-block - get latest block

USAGE:
   HayekTool get-latest-block [command options] [arguments...]

OPTIONS:
   --host value  set host url
   --help, -h    show help (default: false)
```

获取最新上链的Block:

```
$ HayekTool get-latest-block
{
        "number": "0x84e6",
        "hash": "0x9a63cd7147f3e84ecc69d770cf0668c27721ac63bb44f6643eb1fba385e30fef",
        "parentHash": "0x09158e3a6959e2a59aae97674b11399ff77888806fd39fc09da47b914d97542d",
        "nonce": "0x048110be000162ca",
        "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        "transactionsRoot": "0xadd3c61181aa174ba23030ad52a070c603082f6902130a9d0f4d41fa791a277b",
        "stateRoot": "0x74802d7860ab754af02ad03db901010e1a1f9c4b4ae0fe4d873f29cb1517809a",
        "miner": "0xec4dbd592f002f17aea403c2b65fd88f04589cbf",
        "difficulty": "0x776726",
        "totalDifficulty": "0x393f9b331a",
        "extraData": "0xd78201008467716b6388676f312e31332e34856c696e7578",
        "size": "0x290",
        "gasLimit": "0x7a1200",
        "gasUsed": "0x5208",
        "timestamp": "0x5f03b4d8",
        "transactions": [
                {
                        "gas": "0x5208",
                        "gasPrice": "0x174877e552",
                        "hash": "0xb52040b5ac63ddeddc65d303d307f9610965398de3cf6bfcc31871673127082b"
                }
        ],
        "uncles": [],
        "sealFields": null
}
```

## 获取交易信息

交易的hash在Block的transactions字段.

```
$ HayekTool get-tx -hash=0xb52040b5ac63ddeddc65d303d307f9610965398de3cf6bfcc31871673127082b
{
        "transactionHash": "0xb52040b5ac63ddeddc65d303d307f9610965398de3cf6bfcc31871673127082b",
        "transactionIndex": "0x0",
        "blockNumber": "0x84e6",
        "blockHash": "0x9a63cd7147f3e84ecc69d770cf0668c27721ac63bb44f6643eb1fba385e30fef",
        "cumulativeGasUsed": "0x5208",
        "gasUsed": "0x5208",
        "contractAddress": "",
        "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "status": "0x1"
}
```
