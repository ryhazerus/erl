// Package store defines the [Store] interface for rate limit counter backends
// and provides two implementations:
//
//   - [MemoryStore]: fast, in-memory counters that are lost on restart.
//   - [SQLiteStore]: persistent counters backed by a SQLite database.
//
// Custom backends can be created by implementing the [Store] interface.
package store
