package main

import (
	"github.com/aziskebanaran/bvm-core/pkg/wallet"
	"github.com/aziskebanaran/bvm-core/pkg/client"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const CORE_URL = "http://127.0.0.1:8080"

func main() {
    if len(os.Args) < 2 {
        fmt.Println("💡 Gunakan perintah: create, check, atau send")
        return
    }

    // 🚩 PUSAT LOGISTIK TUNGGAL
    const DATA_DIR = "../data"
    walletPath := DATA_DIR + "/node_wallet.json"

    // Pastikan folder markas besar ada
    os.MkdirAll(DATA_DIR, 0755)

    bvm := client.NewBVMClient(CORE_URL)
    command := os.Args[1]

    // Helper sederhana agar case-case di bawah tidak error
    getWalletPath := func() string {
        return walletPath
    }

    switch command {
    case "create":
        newWallet, mnemonic, err := wallet.CreateNewWallet()
        if err != nil {
            fmt.Println("❌ Gagal membuat wallet:", err)
            return
        }

        // Simpan langsung ke jalur pasti
        if err := wallet.SaveWallet(newWallet, walletPath); err != nil {
            fmt.Println("❌ Gagal simpan ke folder data:", err)
            return
        }

        fmt.Println("--------------------------------------------------")
        fmt.Printf("✨ WALLET BERHASIL DIPAHAT!\n")
        fmt.Printf("👤 ADDRESS  : %s\n", newWallet.Address)
        fmt.Printf("📂 DISIMPAN : %s\n", walletPath)
        fmt.Println("--------------------------------------------------")
        fmt.Println("⚠️  CATAT & SIMPAN 12 KATA RAHASIA INI:")
        fmt.Printf("🔑 %s\n", mnemonic)
        fmt.Println("--------------------------------------------------")
        fmt.Println("NB: Mnemonic adalah satu-satunya cara memulihkan saldo emas Anda!")

    // SISA CASE (register, check, send) TETAP SEPERTI ASLINYA...


case "register":
    if len(os.Args) < 3 {
        fmt.Println("💡 Gunakan: register [USERNAME]")
        return
    }
    username := os.Args[2]
    walletFile := getWalletPath()
    senderWallet, _ := wallet.LoadWallet(walletFile)

    // 1. Ambil data jaringan untuk Nonce & Fee
    state, _ := bvm.GetSecureState(senderWallet.Address)

    // 2. Gunakan fungsi konstruktor yang kita buat di x/bvm/types tadi
    // Kita anggap fee registrasi adalah 1 BVM (atau sesuaikan)
    regTx := types.NewRegisterTransaction(senderWallet.Address, username, 100000000, state.Nonce)

    // 3. Tanda tangani
    // (Gunakan logika signing yang sama dengan 'send')
    // Lalu broadcast
    txID, err := bvm.BroadcastTX(regTx)
    if err != nil {
        fmt.Println("❌ Gagal Daftar:", err)
        return
    }
    fmt.Printf("✅ Permintaan Registrasi Terkirim! TXID: %s\n", txID)


case "check":
    walletFile := getWalletPath()
    myWallet, err := wallet.LoadWallet(walletFile)
    if err != nil {
        fmt.Println("❌ Dompet tidak ditemukan.")
        return
    }

    // 🚩 PERBAIKAN: Gunakan GetSecureState agar data 100% sama dengan Node
    state, err := bvm.GetSecureState(myWallet.Address)
    if err != nil {
        fmt.Println("❌ Kernel Offline. Jalankan 'bvm node start' dulu.")
        return
    }

    fmt.Println("---------------------------------------")
    fmt.Printf("👤 ALAMAT : %s\n", state.Address)
    fmt.Printf("💰 SALDO  : %s %s\n", state.BalanceDisplay, state.Symbol) // Pakai string display!
    fmt.Printf("🔢 NONCE  : %d\n", state.Nonce)
    fmt.Printf("📂 FILE   : %s\n", walletFile)
    fmt.Println("---------------------------------------")
    getHistory(myWallet.Address)


        case "send":
                if len(os.Args) < 3 {
                        fmt.Println("💡 Gunakan: send [ALAMAT] [JUMLAH]")
                        return
                }

                var toAddr string
                var amountStr string

                if len(os.Args) >= 4 && (len(os.Args[2]) >= 3 && os.Args[2][:3] == "to=") {
                        for _, arg := range os.Args {
                                if len(arg) > 3 && arg[:3] == "to=" { toAddr = arg[3:] }
                                if len(arg) > 7 && arg[:7] == "amount=" { amountStr = arg[7:] }
                        }
                } else {
                        toAddr = os.Args[2]
                        if len(os.Args) > 3 { amountStr = os.Args[3] }
                }


                // 1. Parsing Float dari Input

    amountFloat, _ := strconv.ParseFloat(amountStr, 64)
    if toAddr == "" || amountFloat <= 0 {
        fmt.Println("❌ Format salah!")
        return
    }

    walletFile := getWalletPath()
    senderWallet, err := wallet.LoadWallet(walletFile)
    if err != nil {
        fmt.Println("❌ Dompet tidak ditemukan"); return
    }

    // 🚩 PERBAIKAN: Ambil Params dari Node agar konversi Atomic akurat
    info, err := bvm.GetNetworkInfo()
    if err != nil {
        fmt.Println("❌ Gagal terhubung ke Node"); return
    }

    // Gunakan fungsi bawaan Sultan: ToAtomic
    amountAtomic := info.Params.ToAtomic(fmt.Sprintf("%.8f", amountFloat))

    // 🚩 PERBAIKAN: JANGAN set senderWallet.Nonce secara manual di sini!
    // Biarkan SignAndPack yang menghitung (State + Mempool) agar tidak bentrok.

    fmt.Printf("🛰️  Menyiapkan Transaksi untuk %s...\n", toAddr)
    signedTx, err := senderWallet.SignAndPack(bvm, toAddr, amountAtomic, "BVM", "Transfer via BVM-Wallet")
    if err != nil {
        fmt.Printf("❌ Gagal Sign: %v\n", err)
        return
    }

    fmt.Printf("🚀 Mengirim %.8f BVM ke %s...\n", amountFloat, toAddr)
    txID, err := bvm.BroadcastTX(signedTx)
    if err != nil {
        fmt.Printf("❌ Jaringan Menolak: %v\n", err)
        return
    }

    // Simpan wallet (untuk update Nonce lokal jika perlu)
    wallet.SaveWallet(senderWallet, walletFile)

    fmt.Println("--------------------------------------------------")
    fmt.Printf("✅ BERHASIL! TX ID: %s\n", txID)
    fmt.Println("--------------------------------------------------")

  }
}


func getHistory(address string) {
    // 🚩 Tips: Gunakan endpoint yang sudah kita rapikan di API
    resp, err := http.Get(CORE_URL + "/api/history?address=" + address)
    if err != nil || resp.StatusCode != 200 {
        fmt.Println("⚠️ Gagal mengambil riwayat."); return
    }
    defer resp.Body.Close()

    var history []struct {
        Height int               `json:"height"`
        Time   int64             `json:"time"`
        Tx     types.Transaction `json:"tx"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&history); err != nil { return }
    if len(history) == 0 {
        fmt.Println("\n📭 Belum ada riwayat transaksi."); return
    }

    fmt.Println("\n📜 RIWAYAT TRANSAKSI TERAKHIR:")

    for i := 0; i < len(history) && i < 10; i++ {
        entry := history[i]

        // 🚩 Gunakan logika konversi yang aman
        displayAmount := float64(entry.Tx.Amount) / 1e8 

        // Tentukan apakah ini IN (Masuk) atau OUT (Keluar)
        icon := "📥"
        if entry.Tx.From == address { icon = "📤" }

        displayAddr := entry.Tx.To
        if entry.Tx.From == address {
            displayAddr = entry.Tx.To
        } else {
            displayAddr = entry.Tx.From
        }

        fmt.Printf("%s [%d] %-12s: %.8f BVM (%s)\n",
            icon,
            entry.Height,
            displayAddr[:12], // Potong address agar rapi
            displayAmount,
            time.Unix(entry.Time, 0).Format("02 Jan 15:04"),
        )
    }
}
