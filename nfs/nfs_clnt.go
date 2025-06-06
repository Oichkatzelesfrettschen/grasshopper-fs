package nfs

import (
	"strconv"

	"github.com/goose-lang/primitive/disk"
	"github.com/mit-pdos/go-nfsd/fh"
	"github.com/mit-pdos/go-nfsd/nfstypes"
)

// NfsClient wraps an in-process NFS server for simple unit tests.
type NfsClient struct {
	srv *Nfs
}

// MkNfsClient creates a new client backed by an in-memory disk of size sz.
func MkNfsClient(sz uint64) *NfsClient {
	d := disk.NewMemDisk(sz)
	return &NfsClient{
		srv: MakeNfs(d),
	}
}

// Shutdown stops the underlying server and releases resources.
func (clnt *NfsClient) Shutdown() {
	clnt.srv.ShutdownNfs()
}

// Crash forces the server to terminate without a graceful shutdown.
func (clnt *NfsClient) Crash() {
	clnt.srv.Crash()
}

// CreateOp issues an NFS CREATE request.
func (clnt *NfsClient) CreateOp(fh nfstypes.Nfs_fh3, name string) nfstypes.CREATE3res {
	where := nfstypes.Diropargs3{Dir: fh, Name: nfstypes.Filename3(name)}
	how := nfstypes.Createhow3{}
	args := nfstypes.CREATE3args{Where: where, How: how}
	attr := clnt.srv.NFSPROC3_CREATE(args)
	return attr
}

// LookupOp performs an NFS LOOKUP request.
func (clnt *NfsClient) LookupOp(fh nfstypes.Nfs_fh3, name string) *nfstypes.LOOKUP3res {
	what := nfstypes.Diropargs3{Dir: fh, Name: nfstypes.Filename3(name)}
	args := nfstypes.LOOKUP3args{What: what}
	reply := clnt.srv.NFSPROC3_LOOKUP(args)
	return &reply
}

// GetattrOp fetches attributes for a file handle.
func (clnt *NfsClient) GetattrOp(fh nfstypes.Nfs_fh3) *nfstypes.GETATTR3res {
	args := nfstypes.GETATTR3args{Object: fh}
	attr := clnt.srv.NFSPROC3_GETATTR(args)
	return &attr
}

// WriteOp sends an NFS WRITE request.
func (clnt *NfsClient) WriteOp(fh nfstypes.Nfs_fh3, off uint64, data []byte, how nfstypes.Stable_how) *nfstypes.WRITE3res {
	args := nfstypes.WRITE3args{
		File:   fh,
		Offset: nfstypes.Offset3(off),
		Count:  nfstypes.Count3(len(data)),
		Stable: how,
		Data:   data}
	reply := clnt.srv.NFSPROC3_WRITE(args)
	return &reply
}

// ReadOp issues an NFS READ request.
func (clnt *NfsClient) ReadOp(fh nfstypes.Nfs_fh3, off uint64, sz uint64) *nfstypes.READ3res {
	args := nfstypes.READ3args{
		File:   fh,
		Offset: nfstypes.Offset3(off),
		Count:  nfstypes.Count3(sz)}
	reply := clnt.srv.NFSPROC3_READ(args)
	return &reply
}

// RemoveOp issues an NFS REMOVE request.
func (clnt *NfsClient) RemoveOp(dir nfstypes.Nfs_fh3, name string) nfstypes.REMOVE3res {
	what := nfstypes.Diropargs3{Dir: dir, Name: nfstypes.Filename3(name)}
	args := nfstypes.REMOVE3args{
		Object: what,
	}
	reply := clnt.srv.NFSPROC3_REMOVE(args)
	return reply
}

// MkDirOp sends an NFS MKDIR request.
func (clnt *NfsClient) MkDirOp(dir nfstypes.Nfs_fh3, name string) nfstypes.MKDIR3res {
	where := nfstypes.Diropargs3{Dir: dir, Name: nfstypes.Filename3(name)}
	sattr := nfstypes.Sattr3{}
	args := nfstypes.MKDIR3args{Where: where, Attributes: sattr}
	attr := clnt.srv.NFSPROC3_MKDIR(args)
	return attr
}

// RmDirOp issues an NFS RMDIR request.
func (clnt *NfsClient) RmDirOp(dir nfstypes.Nfs_fh3, name string) nfstypes.RMDIR3res {
	where := nfstypes.Diropargs3{Dir: dir, Name: nfstypes.Filename3(name)}
	args := nfstypes.RMDIR3args{Object: where}
	attr := clnt.srv.NFSPROC3_RMDIR(args)
	return attr
}

// SymLinkOp creates a symlink.
func (clnt *NfsClient) SymLinkOp(dir nfstypes.Nfs_fh3, name string, path nfstypes.Nfspath3) nfstypes.SYMLINK3res {
	where := nfstypes.Diropargs3{Dir: dir, Name: nfstypes.Filename3(name)}
	sattr := nfstypes.Sattr3{}
	symlink := nfstypes.Symlinkdata3{Symlink_attributes: sattr, Symlink_data: path}
	args := nfstypes.SYMLINK3args{Where: where, Symlink: symlink}
	attr := clnt.srv.NFSPROC3_SYMLINK(args)
	return attr
}

// ReadLinkOp reads the target of a symlink.
func (clnt *NfsClient) ReadLinkOp(fh nfstypes.Nfs_fh3) nfstypes.READLINK3res {
	args := nfstypes.READLINK3args{Symlink: fh}
	attr := clnt.srv.NFSPROC3_READLINK(args)
	return attr
}

// CommitOp forces data to stable storage with an NFS COMMIT call.
func (clnt *NfsClient) CommitOp(fh nfstypes.Nfs_fh3, cnt uint64) *nfstypes.COMMIT3res {
	args := nfstypes.COMMIT3args{
		File:   fh,
		Offset: nfstypes.Offset3(0),
		Count:  nfstypes.Count3(cnt)}
	reply := clnt.srv.NFSPROC3_COMMIT(args)
	return &reply
}

// RenameOp issues an NFS RENAME request.
func (clnt *NfsClient) RenameOp(fhfrom nfstypes.Nfs_fh3, from string,
	fhto nfstypes.Nfs_fh3, to string) nfstypes.Nfsstat3 {
	args := nfstypes.RENAME3args{
		From: nfstypes.Diropargs3{Dir: fhfrom, Name: nfstypes.Filename3(from)},
		To:   nfstypes.Diropargs3{Dir: fhto, Name: nfstypes.Filename3(to)},
	}
	reply := clnt.srv.NFSPROC3_RENAME(args)
	return reply.Status
}

// SetattrOp truncates or otherwise sets attributes for a file.
func (clnt *NfsClient) SetattrOp(fh nfstypes.Nfs_fh3, sz uint64) nfstypes.SETATTR3res {
	size := nfstypes.Set_size3{Set_it: true, Size: nfstypes.Size3(sz)}
	attr := nfstypes.Sattr3{Size: size}
	args := nfstypes.SETATTR3args{Object: fh, New_attributes: attr}
	reply := clnt.srv.NFSPROC3_SETATTR(args)
	return reply
}

// ReadDirPlusOp issues a READDIRPLUS request for directory listings.
func (clnt *NfsClient) ReadDirPlusOp(dir nfstypes.Nfs_fh3, cnt uint64) nfstypes.READDIRPLUS3res {
	args := nfstypes.READDIRPLUS3args{Dir: dir, Dircount: nfstypes.Count3(100), Maxcount: nfstypes.Count3(cnt)}
	reply := clnt.srv.NFSPROC3_READDIRPLUS(args)
	return reply
}

// Parallel runs nthread clients in parallel, executing f for each.
func Parallel(nthread int, disksz uint64,
	f func(clnt *NfsClient, dirfh nfstypes.Nfs_fh3) int) int {
	root := fh.MkRootFh3()
	clnt := MkNfsClient(disksz)
	count := make(chan int)
	for i := 0; i < nthread; i++ {
		go func(i int) {
			name := "d" + strconv.Itoa(i)
			clnt.MkDirOp(root, name)
			reply := clnt.LookupOp(root, name)
			if reply.Status != nfstypes.NFS3_OK {
				panic("Parallel")
			}
			dirfh := reply.Resok.Object
			n := f(clnt, dirfh)
			count <- n
		}(i)
	}
	n := 0
	for i := 0; i < nthread; i++ {
		c := <-count
		n += c
	}
	clnt.Shutdown()
	return n
}
