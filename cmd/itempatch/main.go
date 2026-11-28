// Command itempatch applies explicit CSV patches to native Item.SData without
// reserializing unrelated fields. Input and output must be different paths.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"log"
	"os"
	"slices"
	"strconv"
)

func main() {
	input := flag.String("input", "", "source Item.SData")
	output := flag.String("output", "", "new output file")
	source := flag.String("patches", "", "CSV definition patches (optional special, icon and name columns)")
	retire := flag.Bool("retire-obsolete", false, "retire legacy utility definitions and update current tooltips in Item.SData")
	retireCash := flag.Bool("retire-obsolete-cash", false, "remove retired utility products from Cash.SData")
	flag.Parse()
	if *input == "" || *output == "" || *input == *output {
		log.Fatal("distinct input and output paths required")
	}
	raw, err := os.ReadFile(*input)
	if err != nil {
		log.Fatal(err)
	}
	if *retire || *retireCash {
		if *retire == *retireCash || *source != "" {
			log.Fatal("select one catalog update")
		}
		var patched []byte
		if *retire {
			patched, err = retireItemCatalog(raw)
		} else {
			patched, err = retireCashCatalog(raw)
		}
		if err != nil {
			log.Fatal(err)
		}
		writeOutput(*output, patched)
		return
	}
	f, err := os.Open(*source)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	patches := map[sdata.ItemKey]sdata.ItemPatch{}
	if len(rows) == 0 {
		log.Fatal("missing patch header")
	}
	specialOnly := fmt.Sprint(rows[0]) == "[type type_id req_vg special]"
	visual := fmt.Sprint(rows[0]) == "[type type_id req_vg stack description special icon name]"
	if !specialOnly && !visual && fmt.Sprint(rows[0]) != "[type type_id req_vg stack description]" {
		log.Fatal("invalid patch header")
	}
	for _, row := range rows[1:] {
		nums := [4]uint64{}
		for i := 0; i < 4; i++ {
			v, e := strconv.ParseUint(row[i], 10, 16)
			if e != nil {
				log.Fatal(e)
			}
			nums[i] = v
		}
		if nums[0] > 255 || nums[1] > 255 || nums[3] > 255 {
			log.Fatal("native byte field overflow")
		}
		patch := sdata.ItemPatch{ReqVG: uint16(nums[2])}
		if specialOnly {
			special := uint8(nums[3])
			patch.Special = &special
			patch.PreserveDescription = true
		} else {
			patch.Stack = uint8(nums[3])
			patch.Description = row[4]
			if visual {
				patch.Special = optionalPatchByte(row[5])
				patch.Icon = optionalPatchByte(row[6])
				patch.Name = row[7]
			}
		}
		patches[sdata.ItemKey{Type: uint8(nums[0]), TypeID: uint8(nums[1])}] = patch
	}
	patched, err := sdata.PatchItemDefinitions(raw, patches)
	if err != nil {
		log.Fatal(err)
	}
	writeOutput(*output, patched)
	fmt.Printf("patched %d native definitions\n", len(patches))
}

func optionalPatchByte(text string) *uint8 {
	if text == "" {
		return nil
	}
	value, err := strconv.ParseUint(text, 10, 8)
	if err != nil {
		log.Fatal(err)
	}
	b := uint8(value)
	return &b
}

func writeOutput(output string, patched []byte) {
	fout, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()
	if _, err = fout.Write(patched); err != nil {
		log.Fatal(err)
	}
}

func isRetiredItem(id uint32) bool {
	if sdata.RetiredConsumable(id) || id >= 30071 && id <= 30077 {
		return true
	}
	switch id {
	case 100041, 100064, 100066, 100069, 100075, 100077, 100139, 100174:
		return true
	default:
		return false
	}
}

func retireItemCatalog(raw []byte) ([]byte, error) {
	items, err := sdata.ParseItems(raw)
	if err != nil {
		return nil, err
	}
	patches := make(map[sdata.ItemKey]sdata.ItemPatch)
	for key, item := range items {
		switch {
		case isRetiredItem(uint32(key.Type)*1000 + uint32(key.TypeID)):
			patches[key] = sdata.ItemPatch{Remove: true}
		case sdata.InertItemBox(uint32(key.Type)*1000+uint32(key.TypeID)) && item.Special == 42:
			special := uint8(0)
			patches[key] = sdata.ItemPatch{Special: &special, ClearSkill: true, Description: "This item has no effect."}
		case key.Type == 100 && item.Special == 33:
			patches[key] = sdata.ItemPatch{ReqVG: item.ReqVG, Description: "Resets learned skills and restores your level's original skill points. Active skill cooldowns continue."}
		}
	}
	plain, err := sdata.PatchItemDefinitions(raw, patches)
	if err != nil {
		return nil, err
	}
	return sdata.EncodeLike(raw, plain)
}

func retireCashCatalog(raw []byte) ([]byte, error) {
	products, err := sdata.ParseCash(raw)
	if err != nil {
		return nil, err
	}
	products = slices.DeleteFunc(products, func(product sdata.CashProduct) bool {
		for _, item := range product.Items {
			if isRetiredItem(item.ItemID) {
				return true
			}
		}
		return false
	})
	plain, err := sdata.EncodeCash(products)
	if err != nil {
		return nil, err
	}
	return sdata.EncodeLike(raw, plain)
}
