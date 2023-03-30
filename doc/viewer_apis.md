# View APIs


## 本地启动服务
服务监听端口: 9998

```bash
git clone https://github.com/bestchains/bc-explorer.git

cd cmd/viewer;
go build main.go

./main -v=5 -dsn='postgres://user:password@ip:port/dbname?sslmode=disable'
```

## 1.浏览器总览页面

### 1.1 获取总览信息

`描述`: 根据选择的通道，得到下面的信息

`接口`: /networks/:network/overview/summary

`返回`:

```json
{
    "blockNumber": "4 uint64 -- 区块高度",
    "txCount": "4 uint64 -- 交易数量",
}
```


### 1.2 分段查询

`描述`: 对交易，区块，按时间范围分段查询

`接口`: /networks/:network/overview/query-by-seg?from=0&interval=5&number=2&type=blocks

`query参数`:
| 参数名称 | 参数描述 | 必填 | 默认值 |
| :--: | :--: | :--: | :--: |
| from | 开始时间, 也就是横轴x=0的时候， | 是 | 当前时间 |
| interval | 周期，单位是秒，一小时传递3600，10分钟传递600. | 是 | 300 |
| number | 查询多少段 | 是 | 5 |
| type | 查询区块(blocks)，或者交易(transactions) | 是 | blocks |


```
按照from=0,interval=5,number=2 理解
|     |      |      |
-5    0     5      10
也就是后端会多计算一段数据[from-interval, from]，如果没有这段计算，图表的开始位置数据是0，与预期是不一致的。
```

`返回`: 

返回值字段对应需要明确
```json
[
    {
        "start": "-5 int64 -- 开始时间",
        "end": "0 int64 -- 结束时间",
        "count": "1 int64 -- 数量",
    },
    {
        "start": 0,
        "end": 5,
        "count": 2,
    },
    {
        "start:"5,
        "end": 10,
        "count": 3,
    }
]
```

### 1.3 每个组织的交易数量

`描述`: 根据creator分组，计算每个组的总数，得到百分比图。

---



## 2. 浏览器区块页面

### 2.1 获取区块列表

`描述`: 获取区块列表, 默认按照时间倒序排列

`接口`: GET /networks/:network/blocks

`query参数`:
| 参数名称 | 参数描述 | 必填 | 默认值 |
| :--: | :--: | :--: | :--: |
| from | 分页开始 | 否 | 0 |
| size | 每页的数量 | 否 | 10 |
| startTime | 开始时间 | 否 |  |
| endTime | 结束时间 | 否 |  |
| blockNumber | 根据区块号搜索，完全匹配 | 否 | |
| blockHash | 根据区块hash搜索，完全匹配 | 否 | | 

`返回`:

```json
{
    "data": [{
        "blockNumber": "block.BlockNumber uint64 -- 区块号",
        "network": "block.Network string -- 通道，格式是<network-name>_<channel-name>",
        "txCount": "block.TxCount int -- 交易数量",
        "blockHash": "block.BlockHash string -- 区块hash",
        "preBlockHash": "block.PreviousBlockHash string -- 上一个区块hash",
        "blockSize": "block.BlockSize int -- 区块大小，单位字节",
        "dataHash": "block.DataHash string -- 数据hash",
        "createdAt": "block.CreatedAt int64 -- 出块时间 秒"
    }],
    "count": "10 int -- 查询总数"
}
```


### 2.1 获取区块详情

`描述`: 获取区块详情

`接口`: GET /networks/:network/blocks/:blockHash

`返回`:

```json
{
    "blockNumber": "block.BlockNumber uint64 -- 区块号",
    "network": "block.Network string -- 通道，格式是<network-name>_<channel-name>",
    "txCount": "block.TxCount int -- 交易数量",
    "blockHash": "block.BlockHash string -- 区块hash",
    "preBlockHash": "block.PreviousBlockHash string -- 上一个区块hash",
    "blockSize": "block.BlockSize int -- 区块大小，单位字节",
    "dataHash": "block.DataHash string -- 数据hash",
    "createdAt": "block.CreatedAt int64 -- 出块时间 秒"
}
```

---

## 3. 浏览器交易页面


### 3.1 获取交易列表
`描述`: 获取交易列表

`接口`: GET /networks/:network/transactions

`query参数`:
| 参数名称 | 参数描述 | 必填 | 默认值 |
| :--: | :--: | :--: | :--: |
| from | 分页开始 | 否 | 0 |
| size | 每页的数量 | 否 | 10 |
| startTime | 开始时间 | 否 |  |
| endTime | 结束时间 | 否 |  |
| blockNumber | 根据区块号搜索，完全匹配 | 否 | |

`返回`:

```json
{
    "data": [{
        "id": "transaction.ID string -- 交易ID，交易Hash",
        "network": "transaction.Network string -- 通道，格式<network-name>_<channel-name>",
        "blockNumber": "transaction.BlockNumber uint64 -- 区块号",
        "createdAt": "transaction.CreatedAt int64 -- 时间 秒",
        "creator": "transaction.Creator string -- 发起者",
        "type": "transaction.Type string -- 类型",
        "chaincodeId": "transaction.ChainCodeid string -- 合约",
        "method": "transaction.Method string -- 合约相关的方法",
        "args": "transaction.Args [string] -- 合约相关参数",
        "validationCode": "transaction.ValidationCode int32 -- 交易验证码 0是有效",
        "payload": "transatcion.Payload []byte -- Payload Proplsal Hash"
    }],
    "count": 1
}
```


### 3.2 获取交易详情

`描述`: 获取交易详情


`接口`: GET /networks/:network/transactions/:txHash

`返回`:

```json
{
    "id": "transaction.ID string -- 交易ID，交易Hash",
    "network": "transaction.Network string -- 通道，格式<network-name>_<channel-name>",
    "blockNumber": "transaction.BlockNumber uint64 -- 区块号",
    "createdAt": "transaction.CreatedAt int64 -- 时间",
    "creator": "transaction.Creator string -- 发起者",
    "type": "transaction.Type string -- 类型",
    "chaincodeId": "transaction.ChainCodeid string -- 合约",
    "method": "transaction.Method string -- 合约相关的方法",
    "args": "transaction.Args [string] -- 合约相关参数",
    "validationCode": "transaction.ValidationCode int32 -- 交易验证码 0是有效",
    "payload": "transaction.Payload []byte -- Payload Proplsal Hash"
}
```

### 3.3 获取由特定组织创建的交易数量

`描述`: 获取由特定组织创建的交易的总数

`接口`: GET /networks/:network/transactionsCount

`返回`:

```json
{
    "data": [{
      "creator": "transaction.Creator string -- 发起者名称",
      "count": "int -- 发起的交易总数"
    }],
    "count": 1
}
```