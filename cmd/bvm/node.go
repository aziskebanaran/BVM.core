package main

import (
	"encoding/json" // 🚩 Tambahkan ini
	"os"            // 🚩 Tambahkan ini
	"fmt"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/node"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x/app"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"github.com/aziskebanaran/bvm-core/x/miner" // 🚩 Impor package miner Sultan
	"github.com/spf13/cobra"
	"time"
	"github.com/aziskebanaran/bvm-core/pkg/client"
)

// getMinerAddress: Mengambil identitas miner secara dinamis dari file wallet
func getMinerAddress(homeDir string) (string, error) {
	// Konsolidasi: Selalu cari di dalam folder data (home)
	walletPath := fmt.Sprintf("%s/node_wallet.json", homeDir)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file %s tidak ditemukan", walletPath)
	}

	data, err := os.ReadFile(walletPath)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file wallet: %v", err)
	}

	var w struct {
		Address string `json:"address"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		return "", fmt.Errorf("format JSON wallet rusak: %v", err)
	}

	if w.Address == "" {
		return "", fmt.Errorf("alamat di dalam wallet kosong")
	}

	return w.Address, nil
}

func startNodeProvider(cmd *cobra.Command, args []string) {
	// 1. Ambil Flag Strategis
	h, _ := cmd.Flags().GetString("home")
	nexusURL, _ := cmd.Flags().GetString("nexus")
	useMiner, _ := cmd.Flags().GetBool("miner")

	logger.Info("SYSTEM", fmt.Sprintf("🏗️  Inisialisasi BVM di %s...", h))

	// 2. DATABASE & STATE (Mengikuti Jalur Home)
	dbPath := fmt.Sprintf("%s/blockchain_db", h)
	store, err := storage.NewLevelDBStore(dbPath, 8)
	if err != nil {
		logger.Error("SYSTEM", "🚨 Gagal membuka database: ", err)
		panic(err)
	}

	// 3. JEMBATAN UDARA (Sinkronisasi Nexus via Flag)
	StartNodeWithSync(nexusURL, store)

	// 4. SETUP APP & KERNEL
	bc := types.NewBlockchain()
	bvmApp := app.NewApp(store, bc)
	bvmApp.Start()

	// 5. MINER INTERNAL (Membaca Wallet dari Home)
	if useMiner {
		go func() {
			time.Sleep(5 * time.Second)
			logger.Success("MINER", "🏗️  Membangunkan Miner Internal Sultan...")

			minerAddr, err := getMinerAddress(h)
			if err != nil {
				logger.Error("MINER", "🚨 KRITIKAL: Miner gagal aktif karena: ", err)
				logger.Error("MINER", fmt.Sprintf("Silakan pastikan node_wallet.json tersedia di %s", h))
				return
			}

			logger.Info("MINER", "👷 Alamat Miner Aktif: "+minerAddr)

			engine := miner.NewMinerEngine(bvmApp.BVMKeeper)
			engine.Start(minerAddr)
		}()
	}

	// 6. JALANKAN FULL NODE
	node.StartFullNode(
		bvmApp.BVMKeeper,
		bvmApp.Mempool,
		bvmApp.P2P,
		store,
		"BVM-Primary-Node-01",
	)
}

func StartNodeWithSync(nexusAddr string, store storage.BVMStore) {
    // 1. Ambil tinggi lokal (Sultan lupa kirim 'store')
    localHeight := getLocalHeight(store)

    // 2. Tanya ke Nexus (Fungsi ini return 2 nilai: hasil & error)
    nexusInfo, err := fetchInfoFromNexus(nexusAddr)
    if err != nil {
        fmt.Printf("⚠️ Gagal kontak Nexus: %v. Melanjutkan mode offline...\n", err)
        return
    }

    // 3. Bandingkan (Pastikan konversi tipe data uint64 cocok)
    if uint64(nexusInfo.Height) > localHeight {
        target := uint64(nexusInfo.Height)
        fmt.Printf("🔄 [BOOTSTRAP] Perangkat tertinggal. Menarik %d blok dari Nexus...\n",
            target - localHeight)

        // 4. Jalankan FastSync (Sultan lupa kirim 'target' dan 'store')
        err := performFastSync(nexusAddr, localHeight, target, store)
        if err != nil {
            fmt.Printf("❌ FastSync Gagal: %v\n", err)
        }
    }
}

// 1. Fungsi untuk cek tinggi blok lokal di database Core
func getLocalHeight(store storage.BVMStore) uint64 {
	var h uint64
	store.Get("m:height", &h)
	return h
}

// 2. Fungsi untuk mengambil info dari Nexus Sultan
func fetchInfoFromNexus(nexusURL string) (*types.NetworkResponse, error) {
	c := client.NewBVMClient(nexusURL)
	return c.GetNetworkInfo()
}

// 3. Fungsi eksekutor FastSync
func performFastSync(nexusURL string, start uint64, target uint64, store storage.BVMStore) error {
	c := client.NewBVMClient(nexusURL)
	return c.FastSync(start, target, store)
}
