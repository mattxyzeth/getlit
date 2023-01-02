package ipfs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"time"

	goCid "github.com/ipfs/go-cid"
	sh "github.com/ipfs/go-ipfs-api"
	w3s "github.com/web3-storage/go-w3s-client"
	w3sHttp "github.com/web3-storage/go-w3s-client/http"
)

var shell *sh.Shell

func Node(ctx context.Context, ch chan string) {
	cmd := exec.CommandContext(ctx, "ipfs", "daemon")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err.Error())
		ch <- err.Error()
		close(ch)
		return
	}

	if serr := cmd.Start(); serr != nil {
		fmt.Println(serr.Error())
		ch <- serr.Error()
		close(ch)
		return
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		ch <- line
	}

	if scerr := scanner.Err(); scerr != nil {
		fmt.Println(scerr.Error())
		ch <- scerr.Error()
	}

	if cerr := cmd.Wait(); cerr != nil {
		fmt.Printf("Command Error: %s\n", cerr.Error())
		ch <- cerr.Error()
	}

	close(ch)
}

func Connect() (*sh.Shell, error) {
	if shell != nil {
		return shell, nil
	}

	shell = sh.NewShell("localhost:5001")

	if _, err := shell.ID(); err != nil {
		fmt.Println("IPFS not found. Starting a new node.")

		ch := make(chan string)
		ctx, cancel := context.WithCancel(context.Background())
		go Node(ctx, ch)

		errorMsg, ok := <-ch

		if !ok {
			cancel()
			return nil, fmt.Errorf(errorMsg)
		}
	} else {
		fmt.Println("IPFS connected to a local node.")
	}

	return shell, nil
}

func Web3Get(cid string, web3Token string) (*w3sHttp.Web3Response, error) {
	ipfs, err := w3s.NewClient(w3s.WithToken(web3Token))
	if err != nil {
		return nil, err
	}

	itemCid, err := goCid.Parse(cid)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*60))

	res, err := ipfs.Get(ctx, itemCid)
	if err != nil {
		cancel()
		fmt.Println("Web3Get Get error")
		return nil, err
	}

	return res, nil
}

func FilesForWeb3Res(res *w3sHttp.Web3Response) (files [][]byte, err error) {
	var fsys fs.FS
	_, fsys, err = res.Files()
	if err != nil {
		return
	}

	err = fs.WalkDir(fsys, "/", func(path string, d fs.DirEntry, err error) error {
		info, _ := d.Info()
		if !info.IsDir() {
			file, err := fsys.Open(path)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			files = append(files, data)
		}

		return err
	})

	return
}

func FileFromWeb3Res(res *w3sHttp.Web3Response, name string) (file []byte, err error) {
	var fsys fs.FS
	var f fs.File
	_, fsys, err = res.Files()
	if err != nil {
		return
	}

	f, err = fsys.Open("/blocks/" + name)
	if err != nil {
		return
	}

	return io.ReadAll(f)
}
