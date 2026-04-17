package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/aziskebanaran/bvm-core/x"
	factorytypes "github.com/aziskebanaran/bvm-core/x/factory/types"
)

func HandleL2Anchor(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var report struct {
            NexusID   string `json:"nexus_id"` // Sesuaikan dengan pengirim
            Height    int64  `json:"height"`   // Sesuaikan dengan pengirim
            Hash      string `json:"hash"`
            Operator  string `json:"operator"`
            Timestamp int64  `json:"timestamp"`
        }

        if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // 🚩 INTEGRASI FACTORY KEEPER
        // Cek apakah Nexus-Alpha ini resmi terdaftar di database Sultan
        nexus, err := k.GetFactory().GetNexus(report.NexusID)
        if err != nil {
            fmt.Printf("🚨 [L1] Ilegal! Nexus %s mencoba melapor tanpa izin.\n", report.NexusID)
            http.Error(w, "Nexus tidak terdaftar!", http.StatusForbidden)
            return
        }

        // Simpan Anchor secara permanen
        _ = k.GetFactory().RecordAnchor(report.NexusID, report.Height)

        fmt.Printf("\n🛰️  [L1] Laporan Sah dari %s (%s)!\n", report.NexusID, nexus.NativeToken)
        fmt.Printf("📦 Progress: Blok L2 #%d terpatri di Mainnet.\n", report.Height)

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "OK", "msg": "Anchor Saved"})
    }
}

func HandleRegisterNexus(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Filter hanya metode POST
        if r.Method != http.MethodPost {
            http.Error(w, "Gunakan POST!", 405)
            return
        }

        // 2. Tangkap kiriman JSON dari user/nexus
        var req struct {
            NexusID     string `json:"nexus_id"`
            Owner       string `json:"owner"`
            NativeToken string `json:"native_token"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "JSON Error", 400)
            return
        }

        // 🚩 3. Rakit paket AppChain sesuai standar factory/types
        newNexus := factorytypes.AppChain{
            ID:          req.NexusID,
            Owner:       req.Owner,
            NativeToken: req.NativeToken,
            ChainType:   "Rollup", // 🚩 Sesuai field ChainType
            StakeAmount: 1000000,   // 🚩 Sesuai field StakeAmount (Contoh: 1 BVM)
            IsActive:    true,      // 🚩 Sesuai field IsActive
            LastHeight:  0,         // Dimulai dari nol
        }


        // 🚩 4. Kirim sebagai SATU argumen ke Keeper
        err := k.GetFactory().RegisterNexus(newNexus) 
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        // 5. Beri laporan sukses
        fmt.Printf("🏢 [FACTORY] Nexus Baru Terdaftar: %s (Owner: %s)\n", req.NexusID, req.Owner)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "Nexus Sah!",
            "id":     req.NexusID,
        })
    }
}
