package nfs

import (
	"github.com/mit-pdos/go-journal/util"
	"github.com/mit-pdos/go-nfsd/fh"
	"github.com/mit-pdos/go-nfsd/nfstypes"

	"log"
)

// MOUNTPROC3_NULL handles the NULL RPC for the mount service.
func (nfs *Nfs) MOUNTPROC3_NULL() {
	util.DPrintf(1, "MOUNT Null\n")
}

// MOUNTPROC3_MNT implements the MNT RPC to mount the root file system.
func (nfs *Nfs) MOUNTPROC3_MNT(args nfstypes.Dirpath3) nfstypes.Mountres3 {
	reply := new(nfstypes.Mountres3)
	util.DPrintf(1, "MOUNT Mount %v\n", args)
	reply.Fhs_status = nfstypes.MNT3_OK
	reply.Mountinfo.Fhandle = fh.MkRootFh3().Data
	return *reply
}

// MOUNTPROC3_UMNT handles unmount requests.
func (nfs *Nfs) MOUNTPROC3_UMNT(args nfstypes.Dirpath3) {
	util.DPrintf(1, "MOUNT Unmount %v\n", args)
}

// MOUNTPROC3_UMNTALL unmounts all file systems.
func (nfs *Nfs) MOUNTPROC3_UMNTALL() {
	log.Printf("Unmountall\n")
}

// MOUNTPROC3_DUMP returns mount table entries.
func (nfs *Nfs) MOUNTPROC3_DUMP() nfstypes.Mountopt3 {
	log.Printf("Dump\n")
	return nfstypes.Mountopt3{P: nil}
}

// MOUNTPROC3_EXPORT returns export options.
func (nfs *Nfs) MOUNTPROC3_EXPORT() nfstypes.Exportsopt3 {
	res := nfstypes.Exports3{
		Ex_dir:    "",
		Ex_groups: nil,
		Ex_next:   nil,
	}
	res.Ex_dir = "/"
	return nfstypes.Exportsopt3{P: &res}
}
