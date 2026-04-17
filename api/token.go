package api

import (
    "encoding/json"
    "net/http"
    "github.com/aziskebanaran/bvm-core/x"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
    banktypes "github.com/aziskebanaran/bvm-core/x/bank/types"
	"time"
	"fmt"
)

func HandleCreateToken(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Metode tidak diizinkan", 405)
            return
        }

        var req struct {
            Symbol      string `json:"symbol"`
            TotalSupply uint64 `json:"total_supply"`
            Owner       string `json:"owner"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Format JSON salah", 400)
            return
        }

        // 1. Cek Aktivasi Fitur di Keeper
        if !k.IsFeatureActive("TOKEN_FACTORY", int64(k.GetLastHeight())) {
            http.Error(w, "TOKEN_FACTORY belum aktif!", 403)
            return
        }

        // 2. Rakit Transaksi (Gunakan Struktur Resmi Sultan)
        payload, _ := json.Marshal(req)
        tx := types.Transaction{
            Type:      "create_token",
            From:      req.Owner,
            Payload:   payload,
            Fee:       1000000, 
            Nonce:     k.GetNextNonce(req.Owner),
            Timestamp: time.Now().Unix(), // 🚩 Sultan butuh ini untuk validasi ID
            Symbol:    "BVM",           // Biaya admin pakai BVM
        }

        // 🚩 SOLUSI: Gunakan GenerateID() dan simpan ke field ID
        tx.ID = tx.GenerateID()

        // 3. GUNAKAN PROSES TRANSAKSI (Gerbang Keamanan Sultan)
        if err := k.ProcessTransaction(tx); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "ACCEPTED",
            "message": "Transaksi cetak token valid & masuk Mempool!",
            "tx_hash": tx.ID, // 👈 Gunakan ID yang sudah di-generate
        })
    }
}


func HandleNexusTokenReport(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            NexusID     string `json:"nexus_id"`
            Symbol      string `json:"symbol"`
            TotalSupply uint64 `json:"total_supply"`
            Owner       string `json:"owner"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "JSON Rusak", 400)
            return
        }

        // 1. Validasi legalitas Nexus via Factory
        _, err := k.GetFactory().GetNexus(req.NexusID)
        if err != nil {
            http.Error(w, "Nexus tidak dikenal!", 403)
            return
        }

        // 2. Suntikkan Metadata ke Bank Mainnet (Hanya Metadata, bukan Saldo)
        // Ini agar Querier Wallet Sultan bisa mengenali simbol ini
        msg := banktypes.MsgCreateToken{
            Symbol:      req.Symbol,
            TotalSupply: req.TotalSupply,
            Owner:       req.Owner,
        }
        // Panggil fungsi CreateToken di Bank Mainnet untuk mendaftarkan simbol
        k.GetBank().HandleMsgCreateToken(msg)

        fmt.Printf("🪙 [NEXUS-FACTORY] Token %s resmi terdaftar di Mainnet via %s\n", req.Symbol, req.NexusID)

        w.WriteHeader(200)
    }
}


func HandleSyncNexusToken(k x.BVMKeeper) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req banktypes.MsgCreateToken
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Data tidak valid", 400)
            return
        }

        // Daftarkan metadata ke bank Mainnet secara NATIVE
        // Ini agar Querier Wallet Sultan mengenali simbol baru ini
        err := k.GetBank().CreateToken(req.Owner, req.Symbol, req.TotalSupply)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        fmt.Printf("📢 [SYNC] Token %s dari Nexus telah Terdaftar!\n", req.Symbol)
        w.WriteHeader(http.StatusOK)
    }
}
