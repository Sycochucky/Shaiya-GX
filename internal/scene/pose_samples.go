package scene

import "github.com/Spelunkern/Shaiya-GX/internal/render"

type poseSampleKey struct {
	clip *Clip
	time float32
}

// PoseSamples shares unchanged loop poses during one creature update pass.
// Actor-owned matrices stay independent for later transitions and attachments.
type PoseSamples struct {
	samples map[poseSampleKey][]render.M4
}

func (p *PoseSamples) Reset() { clear(p.samples) }

func (p *PoseSamples) Sample(c *Character) {
	if c == nil || c.clip == nil || c.oneShot || c.blendLeft > 0 {
		return
	}
	key := poseSampleKey{c.clip, c.Time}
	if matrices, ok := p.samples[key]; ok {
		if c.lastFinalsClip != c.clip || c.lastFinalsTime != c.Time || c.lastFinals == nil {
			c.lastFinals = poseBuffer(c.lastFinals, len(matrices))
			copy(c.lastFinals, matrices)
			c.lastFinalsClip, c.lastFinalsTime = c.clip, c.Time
		}
		return
	}
	if p.samples == nil {
		p.samples = make(map[poseSampleKey][]render.M4)
	}
	p.samples[key] = c.finals()
}
