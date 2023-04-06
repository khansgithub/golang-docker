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

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/gin-gonic/gin"

	"golang.org/x/sync/errgroup"
)

func check() {
	// if r := recover(); r != nil {
	// 	fmt.Println("There was a panic")
	// 	fmt.Println(r)
	// }
}

func setup_router() *gin.Engine {
	router := gin.Default()
	router.POST("/workflow/:workflow_name", post_root)
	return router
}

func post_root(c *gin.Context) {
	defer check()

	output_ch := make(chan string, 1)
	workflow_name := c.Param("workflow_name")
	error_group.Go(func() error {
		return create_python_workflow(workflow_name, output_ch)
	})
	err := error_group.Wait()
	handle_error(err)

	c.Status(http.StatusOK)
}

func create_python_workflow(workflow_name string, output_ch chan string) error {
	defer close(output_ch)
	ctx := context.Background()
	container_configs := NewContainerConfig(workflow_name)
	resp, err := docker_cli.ContainerCreate(
		ctx,
		container_configs.Config,
		container_configs.HostConfig,
		nil, nil,
		container_configs.ContainerName,
	)
	handle_error(err)

	output_ch <- "Container ID: " + resp.ID

	// err = docker_cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	// handle_error(err)

	return err
}

type ContainerConfigs struct {
	Config        *container.Config
	HostConfig    *container.HostConfig
	ContainerName string
}

func NewContainerConfig(workflow_name string) *ContainerConfigs {
	workdir := "/usr/home"
	imageName := "python:latest"
	containerName := workflow_name + "_workflow"
	hostPath := "/clone_dir/" + workflow_name
	containerPath := "/usr/home/" + workflow_name

	config := &container.Config{
		Image:      imageName,
		Cmd:        []string{"python", "-u", workflow_name},
		WorkingDir: workdir,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", hostPath, containerPath)},
	}

	return &ContainerConfigs{config, hostConfig, containerName}
}

func docker_cli_f() {
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

func clone_repo() {
	repoURL := "ssh://git@localhost:2222/srv/repo"
	cloneDir := "/clone_dir"

	password := "123"

	// Create an Auth object with password
	auth := &ssh.Password{
		User:     "git", // Replace with your username
		Password: password,
	}

	// Clone the repository with the provided authentication
	_, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
		URL:  repoURL,
		Auth: auth,
	})

	if err != nil {
		fmt.Println("Failed to clone repository:", err)
		os.Exit(1)
	}
}

var error_group *errgroup.Group
var docker_cli *client.Client

func main() {
	// docker_cli()
	clone_repo()
	var err error
	error_group = new(errgroup.Group)
	docker_cli, err = client.NewClientWithOpts(client.FromEnv)
	handle_error(err)

	r := setup_router()
	r.Run(":8080")
}

func handle_error(err error) {
	if err != nil {
		panic(err)
	}
}
