package sdata

import (
	"encoding/binary"
	"fmt"
)

// ItemPatch changes only the specified native definition fields. All other
// record bytes, including unknown fields, are preserved exactly.
type ItemPatch struct {
	Remove              bool
	ReqVG               uint16
	Description         string
	Name                string
	Icon                *uint8
	Stack               uint8
	Special             *uint8
	ClearSkill          bool
	PreserveDescription bool
}

func PatchItemDefinitions(raw []byte, patches map[ItemKey]ItemPatch) ([]byte, error) {
	plain, err := Decrypt(raw)
	if err != nil {
		return nil, err
	}
	c := &cursor{b: plain}
	groups := c.count(4)
	out := binary.LittleEndian.AppendUint32(nil, uint32(groups))
	found := map[ItemKey]bool{}
	for g := 0; g < groups && c.err == nil; g++ {
		count := c.count(60)
		countOffset := len(out)
		out = binary.LittleEndian.AppendUint32(out, uint32(count))
		kept := uint32(count)
		for i := 0; i < count && c.err == nil; i++ {
			start := c.off
			c.str()
			nameEnd := c.off
			c.str()
			descEnd := c.off
			body := append([]byte(nil), c.take(81)...)
			if c.err != nil {
				return nil, c.err
			}
			key := ItemKey{Type: body[0], TypeID: body[1]}
			patch, ok := patches[key]
			if ok && patch.Remove {
				found[key] = true
				kept--
				continue
			}
			if patch.Name != "" {
				out = binary.LittleEndian.AppendUint32(out, uint32(len(patch.Name)))
				out = append(out, patch.Name...)
			} else {
				out = append(out, plain[start:nameEnd]...)
			}
			if ok {
				found[key] = true
				if patch.PreserveDescription {
					out = append(out, plain[nameEnd:descEnd]...)
				} else {
					out = binary.LittleEndian.AppendUint32(out, uint32(len(patch.Description)))
					out = append(out, patch.Description...)
				}
				if patch.ClearSkill {
					body[32], body[33] = 0, 0 // Range and AttackTime
				}
				if patch.Icon != nil {
					body[3] = *patch.Icon
				}
				if patch.Special != nil {
					body[35] = *patch.Special
				}
				binary.LittleEndian.PutUint16(body[28:30], patch.ReqVG)
				if patch.Stack != 0 {
					body[80] = patch.Stack
				}
			} else {
				out = append(out, plain[nameEnd:descEnd]...)
			}
			out = append(out, body...)
		}
		binary.LittleEndian.PutUint32(out[countOffset:], kept)
	}
	if c.err != nil {
		return nil, c.err
	}
	if c.off != len(plain) {
		return nil, fmt.Errorf("Item.SData trailing data")
	}
	for key := range patches {
		if !found[key] {
			return nil, fmt.Errorf("missing patched item %d/%d", key.Type, key.TypeID)
		}
	}
	if _, err := ParseItems(out); err != nil {
		return nil, err
	}
	return out, nil
}
