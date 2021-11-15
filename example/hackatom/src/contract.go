package src

import (
	"github.com/cosmwasm/cosmwasm-go/std"
	"github.com/cosmwasm/cosmwasm-go/std/types"
)

func Instantiate(deps *std.Deps, env types.Env, info types.MessageInfo, msg []byte) (*types.Response, error) {
	deps.Api.Debug("here we go 🚀")

	initMsg := InitMsg{}
	err := initMsg.UnmarshalJSON(msg)
	if err != nil {
		return nil, err
	}

	// just verify these (later we save like that)
	_, err = deps.Api.CanonicalAddress(initMsg.Verifier)
	if err != nil {
		return nil, err
	}
	_, err = deps.Api.CanonicalAddress(initMsg.Beneficiary)
	if err != nil {
		return nil, err
	}

	state := State{
		Verifier:    initMsg.Verifier,
		Beneficiary: initMsg.Beneficiary,
		Funder:      info.Sender,
	}

	err = SaveState(deps.Storage, &state)
	if err != nil {
		return nil, err
	}
	res := &types.Response{
		Attributes: []types.EventAttribute{{"Let the", "hacking begin"}},
	}
	return res, nil
}

func Migrate(deps *std.Deps, env types.Env, info types.MessageInfo, msg []byte) (*types.Response, error) {
	migrateMsg := MigrateMsg{}
	err := migrateMsg.UnmarshalJSON(msg)
	if err != nil {
		return nil, err
	}

	state, err := LoadState(deps.Storage)
	if err != nil {
		return nil, err
	}
	state.Verifier = migrateMsg.Verifier
	err = SaveState(deps.Storage, state)
	if err != nil {
		return nil, err
	}

	res := &types.Response{Data: []byte("migrated")}
	return res, nil
}

func Execute(deps *std.Deps, env types.Env, info types.MessageInfo, data []byte) (*types.ContractResult, error) {
	msg := HandleMsg{}
	err := msg.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	// we need to find which one is non-empty
	switch {
	case msg.Release != nil:
		return executeRelease(deps, &env, &info)
	case msg.CpuLoop != nil:
		return executeCpuLoop(deps, &env, &info)
	case msg.StorageLoop != nil:
		return executeStorageLoop(deps, &env, &info)
	case msg.MemoryLoop != nil:
		return executeMemoryLoop(deps, &env, &info)
	case msg.AllocateLargeMemory != nil:
		return nil, types.GenericError("Not implemented: AllocateLargeMemory")
	case msg.Panic != nil:
		return executePanic(deps, &env, &info)
	case msg.UserErrorsInApiCalls != nil:
		return nil, types.GenericError("Not implemented: UserErrorInApiCalls")
	default:
		return nil, types.GenericError("Unknown HandleMsg")
	}
}

func executeRelease(deps *std.Deps, env *types.Env, info *types.MessageInfo) (*types.ContractResult, error) {
	state, err := LoadState(deps.Storage)
	if err != nil {
		return nil, err
	}

	if info.Sender != state.Verifier {
		return nil, types.Unauthorized{}
	}
	amount, err := std.QuerierWrapper{deps.Querier}.QueryAllBalances(env.Contract.Address)
	if err != nil {
		return nil, err
	}

	msg := types.NewSubMsg(types.CosmosMsg{
		Bank: &types.BankMsg{
			Send: &types.SendMsg{
				ToAddress: state.Beneficiary,
				Amount:    amount,
			},
		},
	})

	res := &types.Response{
		Attributes: []types.EventAttribute{
			{"action", "release"},
			{"destination", state.Beneficiary},
		},
		Messages: []types.SubMsg{msg},
	}
	return &types.ContractResult{Ok: res}, nil
}

func executeCpuLoop(deps *std.Deps, env *types.Env, info *types.MessageInfo) (*types.ContractResult, error) {
	var counter uint64 = 0
	for {
		counter += 1
		if counter >= 9_000_000_000 {
			counter = 0
		}
	}
	return &types.ContractResult{}, nil
}

func executeMemoryLoop(deps *std.Deps, env *types.Env, info *types.MessageInfo) (*types.ContractResult, error) {
	counter := 1
	data := []int{1}
	for {
		counter += 1
		data = append(data, counter)
	}
	return &types.ContractResult{}, nil
}

func executeStorageLoop(deps *std.Deps, env *types.Env, info *types.MessageInfo) (*types.ContractResult, error) {
	var counter uint64 = 0
	for {
		data := []byte{0, 0, 0, 0, 0, 0, byte(counter / 256), byte(counter % 256)}
		deps.Storage.Set([]byte("test.key"), data)
	}
	return &types.ContractResult{}, nil
}

func executePanic(deps *std.Deps, env *types.Env, info *types.MessageInfo) (*types.ContractResult, error) {
	panic("This page intentionally faulted")
}

func Query(deps *std.Deps, env types.Env, data []byte) (*types.QueryResponse, error) {
	msg := QueryMsg{}
	err := msg.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	// we need to find which one is non-empty
	var res std.JSONType
	switch {
	case msg.Verifier != nil:
		res, err = queryVerifier(deps, &env)
	case msg.OtherBalance != nil:
		res, err = queryOtherBalance(deps, &env, msg.OtherBalance)
	case msg.Recurse != nil:
		err = types.GenericError("Not implemented: Recurse")
	default:
		err = types.GenericError("Unknown QueryMsg")
	}
	if err != nil {
		return nil, err
	}

	// if we got a result above, encode it
	bz, err := res.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return types.BuildQueryResponseBinary(bz), nil

}

func queryVerifier(deps *std.Deps, env *types.Env) (*VerifierResponse, error) {
	state, err := LoadState(deps.Storage)
	if err != nil {
		return nil, err
	}

	return &VerifierResponse{
		Verifier: state.Verifier,
	}, nil
}

func queryOtherBalance(deps *std.Deps, env *types.Env, msg *OtherBalance) (*types.AllBalancesResponse, error) {
	amount, err := std.QuerierWrapper{Querier: deps.Querier}.QueryAllBalances(msg.Address)
	if err != nil {
		return nil, err
	}

	return &types.AllBalancesResponse{
		Amount: amount,
	}, nil
}
