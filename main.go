package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func setup_routes() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})
}

func docker_cli() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	handle_error(err)

	ctx := context.Background()

	workdir := "/usr/home"
	imageName := "python:latest"
	containerName := "foo"
	fileName := "foo.py"
	hostPath, err := os.Getwd()
	hostPath += "/" + fileName
	handle_error(err)
	containerPath := "/usr/home/" + fileName

	// var T = true
	config := &container.Config{
		Image: imageName,
		Cmd:   strings.Split("python -u foo.py", " "),
		// Cmd: strings.Split("bash", " "),
		// Env: []string{"PYTHONUNBUFFERED=1"},
		// Cmd:        []string{"python", "foo.py"},
		WorkingDir: workdir,
		Tty:        true,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", hostPath, containerPath)},
		// Init: &T,
	}

	// config := container.Config{
	// 	Image: "python3",
	// 	Cmd: strings.Split("python foo.py", ""),
	// }
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	handle_error(err)

	fmt.Println("Container ID:", resp.ID)

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	handle_error(err)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
		case err := <-errCh:
			handle_error(err)
		case status := <-statusCh:
			// Print the exit code of the container
			fmt.Printf("Container exited with status %d\n", status.StatusCode)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	handle_error(err)
	defer out.Close()

	_, err = io.Copy(os.Stdout, out)

	handle_error(err)
	// fmt.Print(out_text)
	// if _, err := io.Copy(os.Stdout, out); err != nil {
	// 	panic(err)
	// }
}

func main() {
	docker_cli()
	// setup_routes()
	// http.ListenAndServe(":8080", nil)
}

func handle_error(err error) {
	if err != nil {
		panic(err)
	}
}
