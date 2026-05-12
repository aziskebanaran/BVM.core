package keeper

import (
    "fmt"
)

// CallGovernance: Memanggil kontrak manajemen di folder luar secara internal
func (k *Keeper) CallGovernance(method string, args ...interface{}) (interface{}, error) {
    // 1. Definisikan Alamat Kontrak Governance Utama
    const GovContractAddr = "system_gov_manager"

    // 2. Lakukan Query ke Mesin WASM
    // Jenderal menggunakan k.Wasm (WasmKeeper) yang sudah ada di interface
    result, err := k.Wasm.QueryContract(GovContractAddr, method, args...)
    if err != nil {
        return nil, fmt.Errorf("⚖️ GOV_BRIDGE: Gagal memanggil metode [%s]: %v", method, err)
    }

    return result, nil
}

// ProcessBridgeRelease: Menerima perintah dari Nexus untuk mencairkan koin via Governance
func (k *Keeper) ProcessBridgeRelease(to string, amount uint64, ref string) error {
    fmt.Printf("⛓️ [CORE-BRIDGE] Memproses permintaan pelepasan: %d BVM ke %s\n", amount, to)

    // 1. Panggil Kontrak Governance untuk mengeksekusi pengiriman
    // Kita gunakan CallGovernance yang sudah Jenderal buat di atas
    _, err := k.CallGovernance("release_bridge_asset", to, amount, ref)
    if err != nil {
        return fmt.Errorf("❌ [GOV_REJECTED]: %v", err)
    }

    fmt.Printf("✅ [CORE-BRIDGE] Aset berhasil dilepaskan untuk ref: %s\n", ref)
    return nil
}
