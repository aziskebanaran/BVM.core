package keeper

import (
	"github.com/aziskebanaran/bvm-core/pkg/storage"
)


// Gunakan definisi ini sebagai standar baru Jenderal
func (k *Keeper) DistributeBlockReward(height int64, fees uint64, batch storage.Batch) (uint64, uint64, error) {
    p := k.GetParamsData()
    activeValidators := k.GetValidatorCount()

    // 1. HITUNG SUBSIDI STANDAR
    subsidi := k.GetSubsidiAtHeight(height, activeValidators)

    // 2. HITUNG PEMBAGIAN FEE
    tip, burnFromFee := p.DistributeFee(fees)

    // 🚩 3. LOGIKA BONUS OTOMATIS (SULTAN MODE)
    vaultAddr := "bvmf_market_system_vault"
    vaultBalance := k.GetBalanceBVM(vaultAddr)

    if vaultBalance > 0 {
        blocksLeft := int64(p.HalvingInterval) - (height % int64(p.HalvingInterval))

        if blocksLeft > 0 {
            bonus := vaultBalance / uint64(blocksLeft)

            // HANYA EKSEKUSI JIKA BATCH TERSEDIA (Keamanan State)
            if batch != nil && bonus > 0 {
                // Potong saldo Vault di dalam batch blok
                err := k.SubBalanceBVM(vaultAddr, bonus, batch)
                if err == nil {
                    // Hanya tambahkan ke subsidi jika pemotongan saldo berhasil
                    subsidi += bonus
                }
            }
        }
    }

    // 4. TOTAL HADIAH AKHIR
    minerTotal := subsidi + tip

    return minerTotal, burnFromFee, nil
}

// GetSubsidiAtHeight: Sekarang membagi subsidi murni dengan jumlah validator
func (k *Keeper) GetSubsidiAtHeight(height int64, validatorCount int) uint64 {
    params := k.GetParamsData()

    if params.HalvingInterval <= 0 {
        return params.InitialReward
    }

    numHalvings := height / int64(params.HalvingInterval)
    if numHalvings >= 64 {
        return 0
    }

    // 1. Hitung total subsidi blok (100%)
    totalBlockSubsidi := params.InitialReward >> uint64(numHalvings)

    // 2. 🚩 PEMBAGIAN: Bagi total subsidi dengan jumlah validator aktif
    if validatorCount <= 1 {
        return totalBlockSubsidi
    }

    return totalBlockSubsidi / uint64(validatorCount)
}
