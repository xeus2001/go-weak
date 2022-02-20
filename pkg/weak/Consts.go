package weak

// Note: For a simple weak reference the states DEAD, ALIVE and USE are enough, but for more complicated structures
//       like a concurrent weak hash-map or a concurrent String pool we need options for the states, as we have to
//       handle more situations like concurrent map resize, concurrent insert, delete and more.
const (
	// DEAD is the state when a weak reference is dead.
	DEAD uint32 = 0x00
	// ALIVE is the state when a weak reference is alive and not used.
	ALIVE uint32 = 0x01
	// USE is the state when an exclusive lock is held to the weak reference.
	USE uint32 = 0x02
	// READ is a USE option and set when the weak reference is being read.
	READ uint32 = 0x10
	// WRITE is a USE option and set when the weak reference is being written.
	WRITE uint32 = 0x20
	// VERIFY is a USE option and set when the weak reference is written, but not yet verified for duplicated.
	VERIFY uint32 = 0x30
	// RELOCATE is a USE option that signals that this reference is currently relocated.
	RELOCATE uint32 = 0x40
	// DEPRECATED is a USE and DEAD option that signals that this reference has been relocated into the new table.
	// The uid will hold the new index at which this string can be found in the new table, when the state is USE,
	// for DEAD the uid will always be 0.
	DEPRECATED uint32 = 0x50
	// GC is a USE option and set when the weak reference is garbage collected, this basically is the same as DEAD,
	// just that the slot can't be used right now, but will be available soon and a spin to wait is fine.
	GC uint32 = 0xf0
	// USE_READ is USE with option READ
	USE_READ uint32 = USE | READ
	// USE_WRITE is USE with option WRITE
	USE_WRITE uint32 = USE | WRITE
	// USE_VERIFY is USE with option VERIFY
	USE_VERIFY uint32 = USE | VERIFY
	// USE_RELOCATE ...
	USE_RELOCATE uint32 = USE | RELOCATE
	// USE_DEPRECATED ...
	USE_DEPRECATED uint32 = USE | DEPRECATED
	// USE_GC is USE with option GC
	USE_GC uint32 = USE | GC
	// DEAD_DEPRECATED is set when a dead slot has been relocated (made inactive).
	DEAD_DEPRECATED = DEAD | DEPRECATED

	// GREEN is set when the green table is active.
	GREEN uint32 = 0x00
	// BLUE is set when the blue table is active.
	BLUE uint32 = 0x01
	// OK is the option to signal that nothing need to be done by threads entering the pool.
	OK uint32 = 0x00
	// INIT_RESIZE is set by the first thread that starts the resize, all others need to spin until this state
	// is over. When the current option is CLEANUP, all threads need to spin until a pending cleanup is done.
	INIT_RESIZE uint32 = 0x10
	// DO_RESIZE is the state while the resize is ongoing.
	DO_RESIZE uint32 = 0x20
	// CLEANUP is the state set by the last thread that need to finish the resize and perform a cleanup of the
	// blue or green table. This job will first flip the active table, then set the obsolete table to nil and finally
	// set the option back to OK, before it will finish the thing it originally wanted to do, insert a new string.
	CLEANUP uint32 = 0x30
)
