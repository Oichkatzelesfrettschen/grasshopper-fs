package dir

import (
	"github.com/tchajed/marshal"

	"github.com/mit-pdos/go-journal/common"
	"github.com/mit-pdos/go-journal/util"
	"github.com/mit-pdos/go-nfsd/fstxn"
	"github.com/mit-pdos/go-nfsd/inode"
	"github.com/mit-pdos/go-nfsd/nfstypes"
)

const DIRENTSZ uint64 = 128
const MAXNAMELEN = DIRENTSZ - 16 // uint64 for inum + uint64 for len(name)

type dirEnt struct {
	inum common.Inum
	name string // <= MAXNAMELEN
}

// IllegalName reports whether name is "." or "..".
func IllegalName(name nfstypes.Filename3) bool {
	n := name
	return n == "." || n == ".."
}

// ScanName searches dip for name without using the directory cache.
func ScanName(dip *inode.Inode, op *fstxn.FsTxn, name nfstypes.Filename3) (common.Inum, uint64) {
	if dip.Kind != nfstypes.NF3DIR {
		return common.NULLINUM, 0
	}
	var inum = common.NULLINUM
	var finalOffset uint64 = 0
	for off := uint64(0); off < dip.Size; off += DIRENTSZ {
		data, _ := dip.Read(op.Atxn, off, DIRENTSZ)
		if uint64(len(data)) != DIRENTSZ {
			break
		}
		de := decodeDirEnt(data)
		if de.inum == common.NULLINUM {
			continue
		}
		if de.name == string(name) {
			inum = de.inum
			finalOffset = off
			break
		}
	}
	return inum, finalOffset
}

// AddNameDir writes a directory entry directly to dip at or after lastoff.
func AddNameDir(dip *inode.Inode, op *fstxn.FsTxn, inum common.Inum,
	name nfstypes.Filename3, lastoff uint64) (uint64, bool) {
	var finalOff uint64

	for off := uint64(lastoff); off < dip.Size; off += DIRENTSZ {
		data, _ := dip.Read(op.Atxn, off, DIRENTSZ)
		de := decodeDirEnt(data)
		if de.inum == common.NULLINUM {
			finalOff = off
			break
		}
	}
	if finalOff == 0 {
		finalOff = dip.Size
	}
	de := &dirEnt{inum: inum, name: string(name)}
	ent := encodeDirEnt(de)
	util.DPrintf(5, "AddNameDir # %v: %v %v %v off %d\n", dip.Inum, name, de, ent, finalOff)
	n, _ := dip.Write(op.Atxn, finalOff, DIRENTSZ, ent)
	return finalOff, n == DIRENTSZ
}

// RemNameDir removes a directory entry for name from dip.
func RemNameDir(dip *inode.Inode, op *fstxn.FsTxn, name nfstypes.Filename3) (uint64, bool) {
	inum, off := LookupName(dip, op, name)
	if inum == common.NULLINUM {
		return 0, false
	}
	util.DPrintf(5, "RemNameDir # %v: %v %v off %d\n", dip.Inum, name, inum, off)
	de := &dirEnt{inum: common.NULLINUM, name: ""}
	ent := encodeDirEnt(de)
	n, _ := dip.Write(op.Atxn, off, DIRENTSZ, ent)
	return off, n == DIRENTSZ
}

// IsDirEmpty reports whether dip contains only "." and "..".
func IsDirEmpty(dip *inode.Inode, op *fstxn.FsTxn) bool {
	var empty bool = true

	// check all entries after . and ..
	for off := uint64(2 * DIRENTSZ); off < dip.Size; {
		data, _ := dip.Read(op.Atxn, off, DIRENTSZ)
		de := decodeDirEnt(data)
		if de.inum == common.NULLINUM {
			off = off + DIRENTSZ
			continue
		}
		empty = false
		break
	}
	util.DPrintf(10, "IsDirEmpty: %v -> %v\n", dip, empty)
	return empty
}

// InitDir initializes dip as a directory with entries for "." and its parent.
func InitDir(dip *inode.Inode, op *fstxn.FsTxn, parent common.Inum) bool {
	if !AddName(dip, op, dip.Inum, ".") {
		return false
	}
	return AddName(dip, op, parent, "..")
}

// MkRootDir initializes dip as the filesystem root directory.
func MkRootDir(dip *inode.Inode, op *fstxn.FsTxn) bool {
	if !AddName(dip, op, dip.Inum, ".") {
		return false
	}
	return AddName(dip, op, dip.Inum, "..")
}

const fattr3XDRsize uint64 = 4 + 4 + 4 + // type, mode, nlink
	4 + 4 + // uid, gid
	8 + 8 + // size, used
	8 + // rdev (specdata3)
	8 + 8 + // fsid, fileid
	(3 * 8) // atime, mtime, ctime

// best estimate of entryplus3 size, excluding the file name. This mirrors the
// XDR layout of `entryplus3` in RFC 1813.
const entryplus3Baggage uint64 = 8 + // fileid
	4 + // name length
	8 + // cookie
	4 + fattr3XDRsize + // post_op_attr header + fattr3
	16 + // name_handle
	8 // pointer

// readdirBase accounts for the fixed portion of a READDIR or READDIRPLUS reply
// (directory attributes, cookie verifier, pointer to the first entry, and the
// final EOF flag).
const readdirBase uint64 = 88 + 8 + 4 + 4

func pad4(n int) uint64 {
	if n%4 == 0 {
		return 0
	}
	return uint64(4 - (n % 4))
}

// dirEntrySize returns the number of bytes contributed by a single directory
// entry in the XDR reply, excluding attributes or file handles.
func dirEntrySize(name string) uint64 {
	l := len(name)
	return 8 + 4 + uint64(l) + pad4(l) + 8 + 4
}

// Apply iterates over directory entries starting at start and invokes f for each.
// XXX inode locking order violated
func Apply(dip *inode.Inode, op *fstxn.FsTxn, start uint64,
	dircount uint64, maxcount uint64,
	f func(*inode.Inode, string, common.Inum, uint64)) bool {
	var eof bool = true
	var ip *inode.Inode
	var begin = uint64(start)
	if begin != 0 {
		begin += DIRENTSZ
	}
	// Track the size of the XDR reply. Start with the fixed portion of the
	// READDIRPLUS response as described in RFC 1813.
	var n uint64 = readdirBase
	// Size of the directory portion (fileid, name, cookie, and pointer).
	var dirbytes uint64 = 0
	for off := begin; off < dip.Size; {
		data, _ := dip.Read(op.Atxn, off, DIRENTSZ)
		de := decodeDirEnt(data)
		util.DPrintf(5, "Apply: # %v %v off %d\n", dip.Inum, de, off)
		if de.inum == common.NULLINUM {
			off = off + DIRENTSZ
			continue
		}

		// Lock inode, if this transaction doesn't own it already
		var own bool = false
		if op.OwnInum(de.inum) {
			own = true
			ip = op.GetInodeUnlocked(de.inum)
		} else {
			ip = op.GetInodeInum(de.inum)

		}

		f(ip, de.name, de.inum, off)

		// Release inode early, if this trans didn't own it before.
		if !own {
			op.ReleaseInode(ip)
		}

		off = off + DIRENTSZ
		// dircount only accounts for the directory entry portion
		// (as returned by READDIR), while maxcount includes the full
		// XDR reply with attributes and file handles.
		dirbytes += dirEntrySize(de.name)
		n += entryplus3Baggage + uint64(len(de.name)) + pad4(len(de.name))
		if dirbytes >= dircount || n >= maxcount {
			eof = false
			break
		}
	}
	return eof
}

// ApplyEnts enumerates directory entries without looking up inodes.
func ApplyEnts(dip *inode.Inode, op *fstxn.FsTxn, start uint64, count uint64,
	f func(string, common.Inum, uint64)) bool {
	var eof bool = true
	var begin = uint64(start)
	if begin != 0 {
		begin += DIRENTSZ
	}
	// Track the encoded size of the READDIR reply. Start with the fixed
	// portion of the response (attributes, cookie verifier, pointer, EOF).
	var n uint64 = readdirBase
	for off := begin; off < dip.Size; {
		data, _ := dip.Read(op.Atxn, off, DIRENTSZ)
		de := decodeDirEnt(data)
		util.DPrintf(5, "Apply: # %v %v off %d\n", dip.Inum, de, off)
		if de.inum == common.NULLINUM {
			off = off + DIRENTSZ
			continue
		}
		f(de.name, de.inum, off)

		off = off + DIRENTSZ
		// Each entry contributes its XDR-encoded directory fields.
		n += dirEntrySize(de.name)
		if n >= count {
			eof = false
			break
		}
	}
	return eof
}

// Caller must ensure de.Name fits
func encodeDirEnt(de *dirEnt) []byte {
	enc := marshal.NewEnc(DIRENTSZ)
	enc.PutInt(uint64(de.inum))
	enc.PutInt(uint64(len(de.name)))
	enc.PutBytes([]byte(de.name))
	return enc.Finish()
}

func decodeDirEnt(d []byte) *dirEnt {
	dec := marshal.NewDec(d)
	inum := dec.GetInt()
	l := dec.GetInt()
	name := string(dec.GetBytes(l))
	return &dirEnt{
		inum: common.Inum(inum),
		name: name,
	}
}
