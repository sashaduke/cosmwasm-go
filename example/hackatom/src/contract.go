package src

import (
	"github.com/cosmwasm/cosmwasm-go/std"
)

func Init(deps *std.Deps, env std.Env, info std.MessageInfo, msg []byte) (*std.InitResult, error) {
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
	res := &std.InitResponse{
		Attributes: []std.EventAttribute{{"Let the", "hacking begin"}},
	}
	return &std.InitResult{Ok: res}, nil
}

func Migrate(deps *std.Deps, env std.Env, info std.MessageInfo, msg []byte) (*std.MigrateResult, error) {
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

	res := &std.MigrateResponse{Data: []byte("migrated")}
	return &std.MigrateResult{Ok: res}, nil
}

func Handle(deps *std.Deps, env std.Env, info std.MessageInfo, data []byte) (*std.HandleResult, error) {
	msg := HandleMsg{}
	err := msg.UnmarshalJSON(data)
	if err != nil {
		return nil, err
	}

	// we need to find which one is non-empty
	switch {
	case msg.Release != nil:
		return handleRelease(deps, &env, &info)
	case msg.CpuLoop != nil:
		return handleCpuLoop(deps, &env, &info)
	case msg.StorageLoop != nil:
		return handleStorageLoop(deps, &env, &info)
	case msg.MemoryLoop != nil:
		return handleMemoryLoop(deps, &env, &info)
	case msg.AllocateLargeMemory != nil:
		return nil, std.NewError("Not implemented: AllocateLargeMemory")
	case msg.Panic != nil:
		return handlePanic(deps, &env, &info)
	case msg.UserErrorsInApiCalls != nil:
		return nil, std.NewError("Not implemented: UserErrorInApiCalls")
	default:
		return nil, std.NewError("Unknown HandleMsg")
	}
}

func handleRelease(deps *std.Deps, env *std.Env, info *std.MessageInfo) (*std.HandleResult, error) {
	state, err := LoadState(deps.Storage)
	if err != nil {
		return nil, err
	}

	if info.Sender != state.Verifier {
		return nil, std.NewError("Unauthorized")
	}
	amount, err := std.QuerierWrapper{deps.Querier}.QueryAllBalances(env.Contract.Address)
	if err != nil {
		return nil, err
	}

	msg := []std.CosmosMsg{{
		Bank: &std.BankMsg{
			Send: &std.SendMsg{
				FromAddress: env.Contract.Address,
				ToAddress:   state.Beneficiary,
				Amount:      amount,
			},
		},
	}}

	res := &std.HandleResponse{
		Attributes: []std.EventAttribute{
			{"action", "release"},
			{"destination", state.Beneficiary},
		},
		Messages: msg,
	}
	return &std.HandleResult{Ok: res}, nil
}

func handleCpuLoop(deps *std.Deps, env *std.Env, info *std.MessageInfo) (*std.HandleResult, error) {
	var counter uint64 = 0
	for {
		counter += 1
		if counter >= 9_000_000_000 {
			counter = 0
		}
	}
	return &std.HandleResult{}, nil
}

func handleMemoryLoop(deps *std.Deps, env *std.Env, info *std.MessageInfo) (*std.HandleResult, error) {
	counter := 1
	data := []int{1}
	for {
		counter += 1
		data = append(data, counter)
	}
	return &std.HandleResult{}, nil
}

func handleStorageLoop(deps *std.Deps, env *std.Env, info *std.MessageInfo) (*std.HandleResult, error) {
	var counter uint64 = 0
	for {
		data := []byte{0, 0, 0, 0, 0, 0, byte(counter / 256), byte(counter % 256)}
		deps.Storage.Set([]byte("test.key"), data)
	}
	return &std.HandleResult{}, nil
}

func handlePanic(deps *std.Deps, env *std.Env, info *std.MessageInfo) (*std.HandleResult, error) {
	panic("This page intentionally faulted")
}

func Query(deps *std.Deps, env std.Env, data []byte) (*std.QueryResponse, error) {
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
		err = std.NewError("Not implemented: Recurse")
	default:
		err = std.NewError("Unknown QueryMsg")
	}
	if err != nil {
		return nil, err
	}

	// if we got a result above, encode it
	bz, err := res.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return std.BuildQueryResponseBinary(bz), nil

}

func queryVerifier(deps *std.Deps, env *std.Env) (*VerifierResponse, error) {
	state, err := LoadState(deps.Storage)
	if err != nil {
		return nil, err
	}

	return &VerifierResponse{
		Verifier: state.Verifier,
	}, nil
}

func queryOtherBalance(deps *std.Deps, env *std.Env, msg *OtherBalance) (*std.AllBalancesResponse, error) {
	amount, err := std.QuerierWrapper{Querier: deps.Querier}.QueryAllBalances(msg.Address)
	if err != nil {
		return nil, err
	}

	return &std.AllBalancesResponse{
		Amount: amount,
	}, nil
}
