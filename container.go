package main

import (
	"archive/tar"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"time"
	"io"
	"os"
	"bytes"
	"io/ioutil"
	"sync"
	"flag"
	"strconv"
)

var containersSet map[int]string

func main() {
	testnum := flag.Int("num",0,"num")
	flag.Parse()
	containersSet = make(map[int]string)

	wg := sync.WaitGroup{}
	for i:=0;i<*testnum;i++ {
		wg.Add(1)
		go create(i,&wg)
		//respId := create()
		//containersSet[i]=respId
	}
	wg.Wait()
	fmt.Println(containersSet)
	wg = sync.WaitGroup{}
	for i:=0;i<*testnum;i++ {
		wg.Add(1)
		go do(&wg,i,containersSet[i])
	}
	wg.Wait()
	fmt.Println("do done.........")
	ctx := context.Background()
	cli, err := client.NewEnvClient()

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Print("Stopping container ", container.ID[:10], "... ")
		if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
			panic(err)
		}
		fmt.Println("Success")
	}

}

func CopyFromContainer(filePath,destPath,containerId string,ctx context.Context, cli *client.Client) {
	returnoutput,out ,_ := cli.CopyFromContainer(ctx,containerId,filePath)
	defer returnoutput.Close()
	fmt.Println(out)

	tr := tar.NewReader(returnoutput)
	_, err := tr.Next()
	if err != nil {
		fmt.Println(err)
	}

	file, err2 := os.Create(destPath)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer file.Close()

	_, err = io.Copy(file, tr)
	if err != nil {
		fmt.Println(err)
	}
}


func SendToContainer(filePath ,destPath,containerId string,ctx context.Context,cli *client.Client) {
	code,e1 := ioutil.ReadFile(filePath)
	if e1 != nil {
		fmt.Println(e1)
	}

	buf_code:=bytes.NewBuffer(code)

	buf0 := new(bytes.Buffer)
	tw := tar.NewWriter(buf0)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: filePath,
		Size: int64(buf_code.Len()),
	}
	err5 := tw.WriteHeader(tarHeader)
	if err5 != nil {
		fmt.Println(err5)
	}
	_, err := tw.Write(buf_code.Bytes())
	if err != nil {
		fmt.Println(err)
	}

	tarreader := bytes.NewReader(buf0.Bytes())

	err1 := cli.CopyToContainer(ctx,containerId,destPath,tarreader, types.CopyToContainerOptions{
		    AllowOverwriteDirWithFile:true,	
	})
	if err1 != nil{
		fmt.Println(err1)	
	}
}

func create(i int,wg *sync.WaitGroup) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := "beta5ubuntu"
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd: []string{"/bin/bash"},
		Tty: true,
		//AttachStdin:true,   
		AttachStdout:true, 
		AttachStderr:true,
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)
	containersSet[i]=resp.ID
	wg.Done()
	//return resp.ID
}

func do(wg *sync.WaitGroup,i int,respID string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

/*
	imageName := "beta5ubuntu"
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd: []string{"/bin/bash"},
		Tty: true,
		//AttachStdin:true,   
		AttachStdout:true, 
		AttachStderr:true,
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)

*/
//	SendToContainer("code.cpp" ,"/tmp/",resp.ID, ctx,cli) 
	SendToContainer("code.cpp" ,"/tmp/",respID, ctx,cli) 

	respexec,err := cli.ContainerExecCreate(ctx,respID,types.ExecConfig{
		//Tty:true,
		//AttachStdin:true,
		//Detach:true, AttachStdout:true, AttachStderr:true,
		Cmd: []string{"sh","/tmp/complie.sh"},
	})

	if err != nil {
		panic(err)
	}

	resprunexec,err := cli.ContainerExecAttach(ctx,respexec.ID,types.ExecStartCheck{
		Tty:true,
	})
	if err != nil {
		fmt.Println(err)
	}

	io.Copy(os.Stdout,resprunexec.Reader)
	fmt.Println("finish exec")
	
	respp,_ := cli.ContainerExecInspect(ctx,respexec.ID)
	fmt.Println(respp.ExitCode)
	
	SendToContainer("in.txt" ,"/tmp/",respID, ctx,cli) 

	time.Sleep(1*time.Second)
	respexecruncode,err := cli.ContainerExecCreate(ctx,respID,types.ExecConfig{
		//Tty:true,
		//AttachStdin:true,
		//Detach:true,
	    	AttachStdout:true,
		AttachStderr:true,
		Cmd: []string{"sh","/tmp/do.sh"},
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	resprunexecruncode,err := cli.ContainerExecAttach(ctx,respexecruncode.ID,types.ExecStartCheck{
		Tty:true,
	})
	if err != nil {
		fmt.Println(err)
	}

	io.Copy(os.Stdout,resprunexecruncode.Reader)

	resppp,_ := cli.ContainerExecInspect(ctx,respexecruncode.ID)
	fmt.Println(resppp.ExitCode)

	fmt.Println("finish exec")

	copytoPath := "/home/go_workplace/test_container/tmp/output" + strconv.Itoa(i) + ".txt"
	fmt.Println(copytoPath)
	CopyFromContainer("/tmp/out.txt",copytoPath,respID,ctx,cli)

	//time.Sleep(1*time.Second)

	wg.Done()
}

/*
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Print("Stopping container ", container.ID[:10], "... ")
		if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
			panic(err)
		}
		fmt.Println("Success")
	}

*/
