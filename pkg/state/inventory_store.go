package state

import (
    "github.com/aziskebanaran/bvm-lib/game"
    "github.com/aziskebanaran/bvm-lib/storage"
    "fmt"
)

type InventoryManager struct {
    // 🚩 PERBAIKAN: Hapus tanda bintang (*). BVMStore adalah interface.
    db storage.BVMStore 
}

// NewInventoryManager: Tambahkan inisialisasi agar aman
func NewInventoryManager(db storage.BVMStore) *InventoryManager {
    return &InventoryManager{db: db}
}

// SaveInventory: Simpan tas ke LevelDB
func (m *InventoryManager) SaveInventory(inv game.Inventory) error {
    key := fmt.Sprintf("inv:%s", inv.OwnerAddress)
    // Sekarang m.db.Put akan terbaca dengan benar
    return m.db.Put(key, inv)
}

// LoadInventory: Ambil tas dari LevelDB
func (m *InventoryManager) LoadInventory(playerAddr string) (*game.Inventory, error) {
    key := fmt.Sprintf("inv:%s", playerAddr)
    var inv game.Inventory

    // Sekarang m.db.Get akan terbaca dengan benar
    err := m.db.Get(key, &inv)
    if err != nil {
        return nil, err
    }
    return &inv, nil
}
