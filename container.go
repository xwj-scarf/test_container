package main

import (
	//"strings"
	"archive/tar"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"time"
	"io"
	"os"
	//"bufio"
	"bytes"
	"io/ioutil"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := "beta7_ubuntu"
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


	respexec,err := cli.ContainerExecCreate(ctx,resp.ID,types.ExecConfig{
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
	//fmt.Println(respp.Running)

    //fa song input.txt
	//	 fmt.Println ("Open")
	//err3 := tar.NewReader(file)
	//if err3 != nil {
	//	fmt.Println(err)
	//}

	b,e := ioutil.ReadFile("input.txt")
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(b)
	fmt.Println(string(b))

	buf1:=bytes.NewBuffer(b)

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: "in.txt",
		Size: int64(buf1.Len()),
	}
	err5 := tw.WriteHeader(tarHeader)
	if err5 != nil {
		fmt.Println(err5)
	}
	_, err = tw.Write(buf1.Bytes())
	if err != nil {
		fmt.Println(err)
	}

	tarreader := bytes.NewReader(buf.Bytes())

	fmt.Println(resp.ID)
	err1 := cli.CopyToContainer(ctx,resp.ID,"/tmp/",tarreader, types.CopyToContainerOptions{
		    AllowOverwriteDirWithFile:true,	
	})
	if err1 != nil{
		fmt.Println(err1)	
	}

	time.Sleep(2*time.Second)
	respexecruncode,err := cli.ContainerExecCreate(ctx,resp.ID,types.ExecConfig{
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
	//fmt.Println(respexecruncode.Code)
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
	
	resppruncode,_ := cli.ContainerExecInspect(ctx,respexecruncode.ID)
	fmt.Println(resppruncode.ExitCode)

	returnoutput,out ,_ := cli.CopyFromContainer(ctx,resp.ID,"/tmp/out.txt")
	defer returnoutput.Close()
	fmt.Println(out)

	tr := tar.NewReader(returnoutput)
	_, err = tr.Next()
	if err != nil {
		fmt.Println(err)
	}

	file, err2 := os.Create("output.txt")
	if err2 != nil {
		fmt.Println(err2)
	}
	defer file.Close()

	_, err = io.Copy(file, tr)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(10*60*time.Second)

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

