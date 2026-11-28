package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Economy struct {
	WeaponEnchantPercent, ArmorEnchantPercent uint32
	VendorBuyPercent, VendorSellPercent       uint32
	RerollGold                                uint32
}

func DefaultEconomy() Economy {
	return Economy{WeaponEnchantPercent: 100, ArmorEnchantPercent: 100, VendorBuyPercent: 100, VendorSellPercent: 100}
}

func (e Economy) Validate() error {
	if e.WeaponEnchantPercent > 10000 || e.ArmorEnchantPercent > 10000 ||
		e.VendorBuyPercent > 1000 || e.VendorSellPercent > 1000 || e.RerollGold > 2000000000 {
		return errors.New("economy values exceed the allowed range")
	}
	return nil
}

func economyPath(root string) string { return filepath.Join(root, ".local", "economy.json") }

func ReadEconomy(root string) (Economy, error) {
	e := DefaultEconomy()
	raw, err := os.ReadFile(economyPath(root))
	if errors.Is(err, os.ErrNotExist) {
		return e, nil
	}
	if err != nil {
		return e, err
	}
	if err = json.Unmarshal(raw, &e); err != nil {
		return e, err
	}
	return e, e.Validate()
}

func SaveEconomy(root string, e Economy) error {
	if err := e.Validate(); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return writeRuntimeFile(economyPath(root), string(raw)+"\n")
}
