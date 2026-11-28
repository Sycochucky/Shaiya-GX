package proto

import (
	"encoding/binary"
	"fmt"
)

const (
	GXEconomySettings Opcode = 0x7f5e
	GXEconomyAction   Opcode = 0x7f5f
)

type EconomySettings struct {
	Revision                uint32
	BuyPercent, SellPercent uint32
	RerollGold              uint32
	Rejected                bool
}

func EconomySettingsPacket(q EconomySettings) []byte {
	w := NewWriter(GXEconomySettings)
	w.U32(q.Revision)
	w.U32(q.BuyPercent)
	w.U32(q.SellPercent)
	w.U32(q.RerollGold)
	w.Bool(q.Rejected)
	return w.Bytes()
}

func DecodeEconomySettings(p []byte) (EconomySettings, error) {
	if len(p) != 19 || Op(p) != GXEconomySettings {
		return EconomySettings{}, fmt.Errorf("invalid economy settings")
	}
	r := NewReader(p)
	settings := EconomySettings{Revision: r.U32(), BuyPercent: r.U32(), SellPercent: r.U32(), RerollGold: r.U32(), Rejected: r.Bool()}
	if settings.Revision == 0 || settings.BuyPercent > 1000 || settings.SellPercent > 1000 {
		return EconomySettings{}, fmt.Errorf("invalid economy values")
	}
	return settings, r.Err()
}

func EconomyActionPacket(revision, amount uint32, original []byte) []byte {
	out := binary.LittleEndian.AppendUint16(nil, uint16(GXEconomyAction))
	out = binary.LittleEndian.AppendUint32(out, revision)
	out = binary.LittleEndian.AppendUint32(out, amount)
	return append(out, original...)
}

func DecodeEconomyAction(p []byte) (revision, amount uint32, original []byte, err error) {
	if len(p) < 16 || len(p) > 21 || Op(p) != GXEconomyAction {
		return 0, 0, nil, fmt.Errorf("invalid economy action length")
	}
	return binary.LittleEndian.Uint32(p[2:6]), binary.LittleEndian.Uint32(p[6:10]), p[10:], nil
}

func ScaleVendorUnitPrice(authored, percent uint32) uint32 {
	return uint32(min(uint64(authored)*uint64(percent)/100, uint64(^uint32(0))))
}
