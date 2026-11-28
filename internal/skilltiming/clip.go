package skilltiming

import "time"

// ClipFloor returns the dedicated instant animation duration for skillID.
// The caller uses the active definition to decide whether the cast path applies.
// Weapon swings and skills without a generated clip return zero.
func ClipFloor(skillID uint16) time.Duration {
	return time.Duration(clipMilliseconds[skillID]) * time.Millisecond
}
