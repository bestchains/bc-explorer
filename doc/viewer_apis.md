# View APIs


## 本地启动服务
服务监听端口: 9998

```bash
git clone https://github.com/bestchains/bc-explorer.git

cd cmd/viewer;
go build main.go

./main -v=5 -dsn='postgres://user:password@ip:port/dbname?sslmode=disable'
```

## ~~1.浏览器总览页面(暂时不用看)~~

<details>
### 1.1 获取接触总览信息

`描述`: 根据选择的通道，得到下面的信息
**需要明确下面的数据如何获取**

- 区块高度 (blockNumber)
- 交易数量 (transaction表的总行数)
- 节点数量 (需要从集群获取?)
- 合约总数 ()

`接口`: /overview

`返回`:

```json
{
    "blockHigh": 4,
    "transaction": 4,
    "nodes": 1,
    "contracts": 1
}
```

### 1.2 节点

`描述`: 返回节点列表
需要获取加入通道的peer节点列表, 需要连接集群

`接口`: /overview/nodes

`返回`: 

返回值字段对应需要明确
```json
{
    "nodeName": "peer-name",
    "nodeType": "??", 
    "org": "",
    "createTime": "2023",
    "status": "??"
}
```


### 1.3 最新区块

`描述`: 返回最新的n条区块数据

`接口`: /overview/latest-block

`返回`: 

```json
{
    "blockHigh": 1,
    "blockHash": "1234",
    "txCount": 1,
    "createTime": "2023"
}
```

### 1.4 数据
`描述`: 展示最近n小时的区块等数据

### 1.5 每个组织的交易数量

`描述`: 根据creator分组，计算每个组的总数，得到百分比图。

---
</details>



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