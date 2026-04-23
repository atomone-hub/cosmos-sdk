package dynamicfee

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	dynamicfeev1 "github.com/cosmos/cosmos-sdk/x/dynamicfee/types"
)

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: dynamicfeev1.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
