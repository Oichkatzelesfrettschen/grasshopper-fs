package super

import (
	"github.com/goose-lang/primitive/disk"

	"github.com/mit-pdos/go-journal/addr"
	"github.com/mit-pdos/go-journal/common"
)

// FsSuper holds computed values describing the on-disk layout.
type FsSuper struct {
	Disk         disk.Disk
	Size         uint64
	nLog         uint64 // including commit block
	NBlockBitmap uint64
	NInodeBitmap uint64
	nInodeBlk    uint64
	Maxaddr      uint64
}

// MkFsSuper builds a super block description for disk d.
func MkFsSuper(d disk.Disk) *FsSuper {
	sz := d.Size()
	nblockbitmap := (sz / common.NBITBLOCK) + 1

	return &FsSuper{
		Disk:         d,
		Size:         sz,
		nLog:         common.LOGSIZE,
		NBlockBitmap: nblockbitmap,
		NInodeBitmap: common.NINODEBITMAP,
		nInodeBlk:    (common.NINODEBITMAP * common.NBITBLOCK * common.INODESZ) / disk.BlockSize,
		Maxaddr:      sz}
}

// MaxBnum returns the maximum block number in the file system.
func (fs *FsSuper) MaxBnum() common.Bnum {
	return common.Bnum(fs.Maxaddr)
}

// BitmapBlockStart returns the block number of the first block bitmap block.
func (fs *FsSuper) BitmapBlockStart() common.Bnum {
	return common.Bnum(fs.nLog)
}

// BitmapInodeStart returns the block number of the first inode bitmap block.
func (fs *FsSuper) BitmapInodeStart() common.Bnum {
	return fs.BitmapBlockStart() + common.Bnum(fs.NBlockBitmap)
}

// InodeStart returns the first block containing inodes.
func (fs *FsSuper) InodeStart() common.Bnum {
	return fs.BitmapInodeStart() + common.Bnum(fs.NInodeBitmap)
}

// DataStart returns the first data block after metadata.
func (fs *FsSuper) DataStart() common.Bnum {
	return fs.InodeStart() + common.Bnum(fs.nInodeBlk)
}

// Block2addr converts a block number to a disk address.
func (fs *FsSuper) Block2addr(blkno common.Bnum) addr.Addr {
	return addr.MkAddr(blkno, 0)
}

// NInode returns the number of inodes in the file system.
func (fs *FsSuper) NInode() common.Inum {
	return common.Inum(fs.nInodeBlk * common.INODEBLK)
}

// Inum2Addr computes the disk address of the given inode number.
func (fs *FsSuper) Inum2Addr(inum common.Inum) addr.Addr {
	return addr.MkAddr(fs.InodeStart()+common.Bnum(uint64(inum)/common.INODEBLK),
		(uint64(inum)%common.INODEBLK)*common.INODESZ*8)
}
