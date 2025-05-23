package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/zeldovich/go-rpcgen/rfc1057"
	"github.com/zeldovich/go-rpcgen/xdr"

	"github.com/mit-pdos/go-nfsd/nfstypes"
)

func pmap_set_unset(prog, vers, port uint32, setit bool) error {
	var cred rfc1057.Opaque_auth
	cred.Flavor = rfc1057.AUTH_NONE

	pmapc, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", rfc1057.PMAP_PORT))
	if err != nil {
		return err
	}
	defer pmapc.Close()
	pmap := rfc1057.MakeClient(pmapc, rfc1057.PMAP_PROG, rfc1057.PMAP_VERS)

	arg := rfc1057.Mapping{
		Prog: prog,
		Vers: vers,
		Prot: rfc1057.IPPROTO_TCP,
		Port: port,
	}

	var res xdr.Bool
	var proc uint32
	if setit {
		proc = rfc1057.PMAPPROC_SET
	} else {
		proc = rfc1057.PMAPPROC_UNSET
	}

	err = pmap.Call(proc, cred, cred, &arg, &res)
	if err != nil {
		return err
	}
	if bool(res) {
		return nil
	}
	if setit {
		return errors.New("failed to set; is program already registered?")
	}
	return errors.New("failed to unset")
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var diskfile = flag.String("disk", "", "disk image")

func main() {
	var name string
	flag.Parse()
	if *diskfile != "" {
		name = *diskfile
	} else {
		fmt.Printf("Argument '-disk <file>' required\n")
		return
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := uint32(listener.Addr().(*net.TCPAddr).Port)

	err = pmap_set_unset(nfstypes.MOUNT_PROGRAM, nfstypes.MOUNT_V3, 0, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not unset mount - is rpcbind service running?\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	err = pmap_set_unset(nfstypes.MOUNT_PROGRAM, nfstypes.MOUNT_V3, port, true)
	if err != nil {
		panic(err)
	}
	defer pmap_set_unset(nfstypes.MOUNT_PROGRAM, nfstypes.MOUNT_V3, port, false)

	pmap_set_unset(nfstypes.NFS_PROGRAM, nfstypes.NFS_V3, 0, false)
	err = pmap_set_unset(nfstypes.NFS_PROGRAM, nfstypes.NFS_V3, port, true)
	if err != nil {
		panic(err)
	}
	defer pmap_set_unset(nfstypes.NFS_PROGRAM, nfstypes.NFS_V3, port, false)

	nfs := MakeNfs(name)

	srv := rfc1057.MakeServer()
	srv.RegisterMany(nfstypes.MOUNT_PROGRAM_MOUNT_V3_regs(nfs))
	srv.RegisterMany(nfstypes.NFS_PROGRAM_NFS_V3_regs(nfs))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept: %v\n", err)
			break
		}

		go srv.Run(conn)
	}
}
