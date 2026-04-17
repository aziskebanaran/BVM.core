package keeper

import (
	"fmt"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x/factory/types"
	banktypes "github.com/aziskebanaran/bvm-core/x/bank/types"
)

type FactoryKeeper struct {
	store storage.BVMStore
	bank  banktypes.BankKeeper
}

func NewFactoryKeeper(s storage.BVMStore, b banktypes.BankKeeper) *FactoryKeeper {
	return &FactoryKeeper{store: s, bank: b}
}

func (k *FactoryKeeper) RegisterNexus(chain types.AppChain) error {
	key := "factory:nexus:" + chain.ID
	// Gunakan pola Sultan: Get(key, &target)
	var existing types.AppChain
	if err := k.store.Get(key, &existing); err == nil {
		return fmt.Errorf("Nexus ID %s sudah eksis", chain.ID)
	}

	// Staking
	err := k.bank.SubBalance(chain.Owner, chain.StakeAmount, "BVM")
	if err != nil {
		return fmt.Errorf("gagal mengunci jaminan: %v", err)
	}

	chain.IsActive = true
	// Gunakan .Put() bukan .Set() sesuai standar storage Sultan
	return k.store.Put(key, chain)
}

// Tambahkan method yang hilang tadi:
func (k *FactoryKeeper) GetNexus(id string) (*types.AppChain, error) {
	key := "factory:nexus:" + id
	var chain types.AppChain
	err := k.store.Get(key, &chain)
	if err != nil { return nil, err }
	return &chain, nil
}

func (k *FactoryKeeper) RecordAnchor(id string, height int64) error {
    nexus, err := k.GetNexus(id)
    if err != nil { return err }
    nexus.LastHeight = height
    return k.store.Put("factory:nexus:"+id, nexus)
}
