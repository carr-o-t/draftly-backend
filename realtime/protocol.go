package realtime

// yjs binary protocol
// reference: https://github.com/yjs/y-websocket/blob/master/src/y-websocket.js

const (
	// outer message types
	MsgSync      = 0
	MsgAwareness = 1

	// inner sync steps, only used inside a MsgSync payload
	SyncStep1  = 0 // client asks for state
	SyncStep2  = 1 // full doc state, reply to step1
	SyncUpdate = 2 // incremental doc update, the only one we save
)

// read a varuint from data at offset and return the value and how many bytes it used
// return ok=false if data ends too early.
func readVarUint(data []byte, offset int) (value int, size int, ok bool) {
	mult := 1
	value = 0
	pos := offset
	for pos < len(data) {
		r := data[pos]
		pos++
		value += int(r&0x7f) * mult
		if r < 0x80 {
			return value, pos - offset, true
		}
		mult *= 128
	}
	return 0, 0, false
}

// check if data is a yjs MsgSync carrying a SyncUpdate,
// the only sync message that contains a doc change should be saved
func syncMsgUpdate(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	outer, outerSize, ok := readVarUint(data, 0)
	if !ok || outer != MsgSync || outerSize >= len(data) {
		return false
	}
	step, _, ok := readVarUint(data, outerSize)
	return ok && step == SyncUpdate
}
