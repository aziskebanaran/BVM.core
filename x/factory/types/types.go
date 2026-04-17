package types

import "encoding/json"

type AppChain struct {
    ID          string `json:"id"`           // Nama unik Nexus
    Owner       string `json:"owner"`        // Pemilik (Address Sultan)
    ChainType   string `json:"chain_type"`   // Rollup / Sidechain
    NativeToken string `json:"native_token"` // Simbol koin di L2
    StakeAmount uint64 `json:"stake_amount"` // Jaminan BVM di L1
    LastHeight  int64  `json:"last_height"`  // Tinggi blok terakhir yang dilaporkan
    IsActive    bool   `json:"is_active"`
}

func (a AppChain) Marshal() []byte {
    data, _ := json.Marshal(a)
    return data
}
