package app

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Spelunkern/Shaiya-GX/internal/data/sdata"
	"github.com/Spelunkern/Shaiya-GX/internal/platform"
	"github.com/Spelunkern/Shaiya-GX/internal/ui"
)

type tooltipContent struct {
	tooltip            []itemDetailLine
	tooltipItem        bool
	tooltipMuted       bool
	tooltipDescription string
	tooltipExtra       tooltipExtras
	tooltipSockets     []tooltipSocket
}

// Shared presentation from Phoenix skill_tooltip.gd / item_tooltip.gd:
// 208px tipbox, 185px body, 12px text, cursor-left placement and gold border.
func (a *App) queueSkillTooltip(key sdata.SkillKey) {
	d, ok := a.skillDefinition(key.ID, key.Level)
	if !ok {
		return
	}
	title := d.Name
	if key.Level != 100 && !strings.HasSuffix(strings.ToLower(strings.TrimSpace(title)), fmt.Sprintf("lv%d", key.Level)) && !strings.HasSuffix(strings.ToLower(strings.TrimSpace(title)), fmt.Sprintf("lv %d", key.Level)) {
		title = fmt.Sprintf("%s Lv%d", title, key.Level)
	}
	rows := []itemDetailLine{{title, [4]float32{1, 1, 0, 1}}}
	if d.Effect != 0 {
		if d.SP > 0 {
			rows = append(rows, itemDetailLine{"SP: " + a.skillCostValue(d.SP), white})
		}
		if d.MP > 0 {
			rows = append(rows, itemDetailLine{"MP: " + a.skillCostValue(d.MP), white})
		}
		if d.ReadyTime > 0 {
			rows = append(rows, itemDetailLine{fmt.Sprintf("Cast time: %g seconds", float64(d.ReadyTime)*.25), white})
		}
		if d.ResetTime > 1 {
			rows = append(rows, itemDetailLine{a.msg.Format(12323, map[string]string{"v": fmt.Sprint(d.ResetTime)}), white})
		}
		if d.AttackRange > 0 {
			rows = append(rows, itemDetailLine{fmt.Sprintf("Range: %d m", d.AttackRange), white})
		}
	}
	if d.Attribute >= 1 && d.Attribute <= 8 {
		rows = append(rows, itemDetailLine{elementName(d.Attribute, a.msg), skillElementColor(d.Attribute)})
	} else if d.Detail == 94 || d.Detail == 95 {
		rows = append(rows, itemDetailLine{elementName(0, a.msg), white})
	}
	warehouseCall := key.ID == warehouseCallSkill && key.Level == 1
	if _, learned := a.learnedSkill(key.ID); key.Level != 100 && !learned && !warehouseCall {
		rows = append(rows, itemDetailLine{a.msg.Format(12324, map[string]string{"v": fmt.Sprint(d.RequiredLevel)}), gray}, itemDetailLine{a.msg.Format(12325, map[string]string{"v": fmt.Sprint(d.Points)}), gray})
	}

	a.tooltipItem = false
	a.tooltipSockets = nil
	a.tooltipDescription = strings.ReplaceAll(d.Description, `\n`, "\n")

	if warehouseCall {
		a.tooltipDescription = warehouseCallDescription
		if seconds := d.KeepSeconds(); seconds > 0 {
			rows = append(rows, itemDetailLine{"Duration: " + ui.BuffDuration(seconds, true), white})
		}
	}
	a.tooltip = rows
}
func (a *App) queueActionTooltip(w, h int) {
	pointer := &platform.Input{MouseX: a.lastMouseX, MouseY: a.lastMouseY}
	if a.barCovered(pointer, w, h) {
		return
	}
	index := a.actionBar.SlotAt(a.lastMouseX, a.lastMouseY, w, h)
	if index < 0 {
		return
	}
	if action, ok := a.localAction(index); ok {
		a.tooltip = []itemDetailLine{{a.subActionName(uint8(action)), white}}
		a.tooltipItem = false
		a.tooltipDescription = ""
		a.tooltipSockets = nil
		return
	}
	q, _ := a.shortcutAt(index)
	if q == nil {
		return
	}
	if q.Bag == 100 {
		if k, ok := a.state.Skills[uint8(q.Number)]; ok {
			a.queueSkillTooltip(sdata.SkillKey{ID: k.ID, Level: k.Level})
		}
		return
	}
	if it, ok := a.state.Inventory[[2]uint8{q.Bag, uint8(q.Number)}]; ok {
		if d := a.itemsItem(it.Type, it.TypeID); d != nil {
			a.drawItemDetails(d, &it, it.Count, 0, 0, w, h, "")
		}
	}
}

type tooltipGlyph struct {
	char  rune
	color [4]float32
}

func tooltipGlyphs(text string, color [4]float32) []tooltipGlyph {
	var out []tooltipGlyph
	colors := map[int][4]float32{0: {1, 0, 1, 1}, 1: {.5, 1, 1, 1}, 2: {0, 1, 0, 1}, 3: {1, 1, 0, 1}, 4: {1, 0, 0, 1}, 5: {.63, .62, .61, 1}, 50: {1, .8, 0, 1}, 51: {.8, 1, .2, 1}}
	base := color
	for len(text) > 0 {
		if strings.HasPrefix(text, "{r}") {
			color = [4]float32{1, 135.0 / 255, 0, 1}
			text = text[3:]
			continue
		}
		if strings.HasPrefix(text, "{/r}") {
			color = base
			text = text[4:]
			continue
		}
		if strings.HasPrefix(text, "{/c}") {
			color = base
			text = text[4:]
			continue
		}
		if strings.HasPrefix(text, "{c") {
			if end := strings.IndexByte(text, '}'); end > 2 {
				if n, err := strconv.Atoi(text[2:end]); err == nil {
					color = white
					if tint, ok := colors[n]; ok {
						color = tint
					}
					text = text[end+1:]
					continue
				}
			}
		}
		r, size := utf8.DecodeRuneInString(text)
		out = append(out, tooltipGlyph{r, color})
		text = text[size:]
	}
	return out
}
func (a *App) drawSharedTooltip(w, h int) { a.drawSharedTooltipAt(w, h, nil) }

func (a *App) drawSharedTooltipAt(w, h int, position *[2]int) {
	for _, dialog := range []*ui.Message{a.accountMessage, a.partyMessage, a.vehicleMessage, a.guildConfirm, a.unionConfirm, a.auctionConfirm, a.guildDuelMessage, a.duelMessage, a.gemConfirm, a.gateConfirm, a.questConfirm, a.exchangeInvite} {
		if dialog != nil && dialog.Visible {
			return
		}
	}
	if a.appearanceEdit != nil || a.shop.quantity != nil || a.exchange.quantity != nil || a.exchange.moneyOpen || a.duel.quantity != nil || a.duel.moneyOpen || a.blacksmith.quantity {
		return
	}
	if a.menuOpen || len(a.tooltip) == 0 || a.state == nil || a.state.Self.Dead || a.skills.drag.active || a.inventory.dragging || (a.wv != nil && a.wv.localActionDrag.active) || a.windowDrag.position != nil || a.actionBar.LayoutActive() {
		return
	}

	f := a.tooltipFont
	if f == nil {
		f = a.font
	}
	heading := a.tooltipHeadingFont
	if heading == nil {
		heading = f
	}
	const width, gap = 208, 8
	tip, comparison := a.tooltipContent, a.tooltipComparison
	layout := tip.layout(f, heading)
	scale := min(float32(1), float32(w)/width, float32(h)/float32(max(1, layout.height)))
	if position == nil && tip.tooltipItem && len(comparison.tooltip) > 0 {
		other := comparison.layout(f, heading)
		scale = min(scale, float32(w)/(2*width+gap), float32(h)/float32(max(1, other.height)))
		left, top, otherLeft := comparisonTooltipOrigins(a.lastMouseX, a.lastMouseY, w, h, max(layout.height, other.height), scale)
		a.drawTooltipContent(tip, layout, f, heading, left, top, scale)
		a.drawTooltipContent(comparison, other, f, heading, otherLeft, top, scale)
		return
	}
	left, top := ui.TooltipOriginFor(a.lastMouseX, a.lastMouseY, width, layout.height, w, h, tip.tooltipItem, scale)
	if position != nil {
		left = max(0, min(position[0], w-int(width*scale)))
		top = max(0, min(position[1], h-int(float32(layout.height)*scale)))
	}
	a.drawTooltipContent(tip, layout, f, heading, left, top, scale)
}

func (tip tooltipContent) layout(f, heading *ui.Font) tooltipLayout {
	return layoutTooltipWithHeading(f, heading, tip.tooltip, tip.tooltipItem, tip.tooltipDescription, len(tip.tooltipSockets), tip.tooltipExtra)
}

func comparisonTooltipOrigins(mx, my, w, h, height int, scale float32) (left, top, otherLeft int) {
	width := int(208 * scale)
	gap := int(8 * scale)
	pair := 2*width + gap
	top = max(0, min(my+10, h-int(float32(height)*scale)))
	if mx >= pair {
		left = min(mx-width, w-width)
		return left, top, left - width - gap
	}
	left = max(0, min(mx+8, w-pair))
	return left, top, left + width + gap
}

func (a *App) drawTooltipContent(tip tooltipContent, layout tooltipLayout, f, heading *ui.Font, left, top int, scale float32) {
	rect := func(x, y, width, height int, color [4]float32) {
		a.sprites.Draw(a.white, float32(left)+float32(x)*scale, float32(top)+float32(y)*scale, float32(width)*scale, float32(height)*scale, 0, 0, 1, 1, color)
	}
	ui.DrawTipFrame(a.sprites, a.white, float32(left), float32(top), 208, float32(layout.height), scale, ui.TipFill, ui.TipBorder)
	for _, run := range layout.runs {
		face := f
		if run.emphasis {
			face = heading
		}
		if tip.tooltipMuted {
			run.color = [4]float32{.6, .6, .6, 1}
		}
		if run.shadow {
			face.DrawScaled(a.sprites, run.text, float32(left)+float32(run.x+1)*scale, float32(top)+float32(run.y+1)*scale, scale, [4]float32{0, 0, 0, 1}, 185)
		}
		face.DrawScaled(a.sprites, run.text, float32(left)+float32(run.x)*scale, float32(top)+float32(run.y)*scale, scale, run.color, 185)
	}
	if tip.tooltipItem {
		i := 0
		for _, row := range [2][]tooltipItemEffect{tip.tooltipExtra.effects, tip.tooltipExtra.immunities} {
			for _, skill := range row {
				if file, uv, ok := skill.icon(); ok {
					if tex := a.assets.Texture("interface/" + file); tex != nil {
						x, y := layout.effectIcons[i][0], layout.effectIcons[i][1]
						a.sprites.Draw(tex, float32(left)+float32(x)*scale, float32(top)+float32(y)*scale, 18*scale, 18*scale, uv.U0, uv.V0, uv.U1, uv.V1, white)
					}
				}
				i++
			}
		}
	}
	if tip.tooltipItem && tip.tooltipExtra.material != nil {
		d := tip.tooltipExtra.material
		if file, uv, ok := ui.ItemIcon(d.Type, d.Icon); ok {
			if tex := a.assets.Texture("interface/" + file); tex != nil {
				a.sprites.Draw(tex, float32(left)+float32(layout.materialX)*scale, float32(top)+float32(layout.materialY)*scale, 14*scale, 14*scale, uv.U0, uv.V0, uv.U1, uv.V1, white)
			}
		}
	}
	for i, socket := range tip.tooltipSockets {
		x, y := 10+i*20, layout.socketY
		if socket.definition == nil {
			rect(x, y, 16, 16, [4]float32{0, 0, 0, .5})
			continue
		}
		d := socket.definition
		if file, uv, ok := ui.ItemIcon(d.Type, d.Icon); ok {
			if tex := a.assets.Texture("interface/" + file); tex != nil {
				if socket.inactive {
					a.sprites.DrawGrayscale(tex, float32(left)+float32(x)*scale, float32(top)+float32(y)*scale, 16*scale, 16*scale, uv.U0, uv.V0, uv.U1, uv.V1, [4]float32{.7, .7, .7, 1})
				} else {
					a.sprites.Draw(tex, float32(left)+float32(x)*scale, float32(top)+float32(y)*scale, 16*scale, 16*scale, uv.U0, uv.V0, uv.U1, uv.V1, white)
				}
			}
		}
	}

}

// Drawing order decides hover ownership, including empty panel backgrounds.
func (a *App) clearTooltipUnderPanel(x, y, w, h int) {
	if a.lastMouseX >= x && a.lastMouseX < x+w && a.lastMouseY >= y && a.lastMouseY < y+h {
		a.tooltip = nil
	}
}
