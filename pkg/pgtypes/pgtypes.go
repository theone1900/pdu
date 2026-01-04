// Package pgtypes defines PostgreSQL data types and structures.
package pgtypes

import "encoding/binary"

// PG_VERSION_NUM defines the PostgreSQL version number.
const PG_VERSION_NUM = 15

// Block size constants
const (
	BLCKSZ      = 8192
	NAMEDATALEN = 64
)

// ItemId flags
const (
	LP_UNUSED   = 0 // unused (should always have lp_len=0)
	LP_NORMAL   = 1 // used (should always have lp_len>0)
	LP_REDIRECT = 2 // HOT redirect (should have lp_len=0)
	LP_DEAD     = 3 // dead, may or may not have storage
)

// Page flags
const (
	PD_HAS_FREE_LINES = 0x0001 // are there any unused line pointers?
	PD_PAGE_FULL      = 0x0002 // not enough free space for new tuple?
	PD_ALL_VISIBLE    = 0x0004 // all tuples on page are visible to everyone
)

// Tuple header info mask bits
const (
	HEAP_HASNULL           = 0x0001 // has null attribute(s)
	HEAP_HASOID            = 0x0008 // has object id
	HEAP_XMAX_IS_MULTI     = 0x1000 // t_xmax is a MultiXactId
	HEAP_NATTS_MASK        = 0x07FF // 11 bits for number of attributes
)

// PageXLogRecPtr represents a pointer to a location in the WAL.
type PageXLogRecPtr struct {
	XLogID  uint32
	XRecOff uint32
}

// ItemIdData represents a line pointer on a page.
type ItemIdData struct {
	LpOff   uint16 // offset to tuple (from start of page)
	LpFlags uint16 // state of line pointer
	LpLen   uint16 // byte length of tuple
}

// HeapPageHeaderData represents the header of a heap page.
type HeapPageHeaderData struct {
	PDLSN             PageXLogRecPtr // LSN: next byte after last byte of xlog record
	PDChecksum        uint16         // checksum
	PDFlags           uint16         // flag bits
	PDLower           uint16         // offset to start of free space
	PDUpper           uint16         // offset to end of free space
	PDSpecial         uint16         // offset to start of special space
	PDPagesizeVersion uint16         // page size and layout version
	PDPruneXID        uint32         // oldest prunable XID, or zero if none
	// PDLinp follows (flexible array)
}

// HeapTupleHeaderData represents the header of a heap tuple.
type HeapTupleHeaderData struct {
	// Fields with transaction information
	THeap struct {
		TXmin      uint32 // inserting xact ID
		TXmax      uint32 // deleting or locking xact ID
		TField3 struct {
			TCid  uint32 // inserting or deleting command ID
			TXvac uint32 // old-style VACUUM FULL xact ID
		}
	}
	
	TCTID     [6]byte // current TID of this or newer tuple
	TInfomask2 uint16  // number of attributes + various flags
	TInfomask  uint16  // various flag bits
	THoff      uint8   // sizeof header incl. bitmap, padding
	TBits      []byte  // bitmap of NULLs (flexible array)
}

// ReadHeapPageHeader reads a HeapPageHeaderData from a byte slice.
func ReadHeapPageHeader(data []byte) HeapPageHeaderData {
	var header HeapPageHeaderData
	
	// Read LSN
	header.PDLSN.XLogID = binary.BigEndian.Uint32(data[0:4])
	header.PDLSN.XRecOff = binary.BigEndian.Uint32(data[4:8])
	
	// Read checksum and flags
	header.PDChecksum = binary.BigEndian.Uint16(data[8:10])
	header.PDFlags = binary.BigEndian.Uint16(data[10:12])
	
	// Read free space pointers
	header.PDLower = binary.BigEndian.Uint16(data[12:14])
	header.PDUpper = binary.BigEndian.Uint16(data[14:16])
	header.PDSpecial = binary.BigEndian.Uint16(data[16:18])
	
	// Read page size version and prune XID
	header.PDPagesizeVersion = binary.BigEndian.Uint16(data[18:20])
	header.PDPruneXID = binary.BigEndian.Uint32(data[20:24])
	
	return header
}

// ReadItemIdData reads an ItemIdData from a byte slice at the given offset.
func ReadItemIdData(data []byte, offset int) ItemIdData {
	var itemId ItemIdData
	
	// Read offset, flags, and length
	itemId.LpOff = binary.BigEndian.Uint16(data[offset : offset+2])
	itemId.LpFlags = binary.BigEndian.Uint16(data[offset+2 : offset+4]) & 0x0003 // only 2 bits for flags
	itemId.LpLen = binary.BigEndian.Uint16(data[offset+2 : offset+4]) >> 2       // remaining 14 bits for length
	
	return itemId
}

// ItemIdHasStorage checks if an ItemId has storage.
func ItemIdHasStorage(itemId ItemIdData) bool {
	return itemId.LpLen != 0
}

// ItemIdIsUsed checks if an ItemId is used.
func ItemIdIsUsed(itemId ItemIdData) bool {
	return itemId.LpFlags != LP_UNUSED
}

// ItemIdGetOffset gets the offset from an ItemId.
func ItemIdGetOffset(itemId ItemIdData) uint16 {
	return itemId.LpOff
}

// ItemIdGetLength gets the length from an ItemId.
func ItemIdGetLength(itemId ItemIdData) uint16 {
	return itemId.LpLen
}

// ItemIdGetFlags gets the flags from an ItemId.
func ItemIdGetFlags(itemId ItemIdData) uint16 {
	return itemId.LpFlags
}
