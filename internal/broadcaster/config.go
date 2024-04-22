package broadcaster

import (
	"fmt"
	"time"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	txclient "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"gitlab.com/distributed_lab/figure/v3"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const accountPrefix = "rarimo"

type config interface {
	comfig.Logger
	pgdb.Databaser
	Broadcasterer
}

type Config struct {
	AirdropCoins types.Coins

	sender        cryptotypes.PrivKey
	senderAddress string
	chainID       string
	txConfig      sdkclient.TxConfig
	txClient      txclient.ServiceClient
	auth          authtypes.QueryClient
	queryLimit    uint64
}

type Broadcasterer interface {
	Broadcaster() Config
}

type broadcasterer struct {
	getter kv.Getter
	once   comfig.Once
}

func New(getter kv.Getter) Broadcasterer {
	return &broadcasterer{
		getter: getter,
	}
}

func (b *broadcasterer) Broadcaster() Config {
	return b.once.Do(func() interface{} {
		var cfg struct {
			AirdropAmount    string `fig:"airdrop_amount,required"`
			CosmosRPC        string `fig:"cosmos_rpc,required"`
			ChainID          string `fig:"chain_id,required"`
			SenderPrivateKey string `fig:"sender_private_key,required"`
			QueryLimit       uint64 `fig:"query_limit"`
		}

		err := figure.Out(&cfg).From(kv.MustGetStringMap(b.getter, "broadcaster")).Please()
		if err != nil {
			panic(fmt.Errorf("failed to figure out broadcaster: %w", err))
		}

		amount, err := types.ParseCoinsNormalized(cfg.AirdropAmount)
		if err != nil {
			panic(fmt.Errorf("broadcaster: invalid airdrop amount: %w", err))
		}

		cosmosRPC, err := grpc.Dial(cfg.CosmosRPC, grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second, // wait time before ping if no activity
			Timeout: 20 * time.Second, // ping timeout
		}))
		if err != nil {
			panic(fmt.Errorf("broadcaster: failed to dial cosmos core rpc: %w", err))
		}

		privateKey, err := hexutil.Decode(cfg.SenderPrivateKey)
		if err != nil {
			panic(fmt.Errorf("broadcaster: sender private key is not a hex string: %w", err))
		}

		sender := &secp256k1.PrivKey{Key: privateKey}
		address, err := bech32.ConvertAndEncode(accountPrefix, sender.PubKey().Address().Bytes())
		if err != nil {
			panic(fmt.Errorf("failed to convert and encode sender address: %w", err))
		}

		queryLimit := uint64(100)
		if cfg.QueryLimit > 0 {
			queryLimit = cfg.QueryLimit
		}

		return Config{
			sender:        sender,
			senderAddress: address,
			chainID:       cfg.ChainID,
			txConfig: authtx.NewTxConfig(
				codec.NewProtoCodec(codectypes.NewInterfaceRegistry()),
				[]signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT},
			),
			txClient:     txclient.NewServiceClient(cosmosRPC),
			auth:         authtypes.NewQueryClient(cosmosRPC),
			AirdropCoins: amount,
			queryLimit:   queryLimit,
		}
	}).(Config)
}
