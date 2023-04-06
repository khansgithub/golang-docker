package main

import (
	"bytes"
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
	router.GET("/workflow/:container_id", get_container_output)
	return router
}

func get_container_output(c *gin.Context) {
	container_id := c.Param("container_id")
	ctx := context.Background()

	out, err := docker_cli.ContainerLogs(ctx, container_id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	defer out.Close()

	bodyBytes, err := io.ReadAll(out)
	handle_error(err)
	bodyBytes = bytes.Trim(bodyBytes, "\x00")

	str := fmt.Sprintf("%s", bodyBytes)
	// var str string
	// for len(bodyBytes) > 0 {
	// 	r, size := utf8.DecodeRune(bodyBytes)
	// 	str += string(r)
	// 	bodyBytes = bodyBytes[size:]
	// }

	// bodyString := string(bodyBytes)
	res := struct {
		Output string `json:"output"`
	}{Output: str}
	c.JSON(http.StatusOK, res)
	return

}

func post_root(c *gin.Context) {
	defer check()

	container_id_ch := make(chan string, 1)
	error_ch := make(chan error, 1)
	workflow_name := c.Param("workflow_name")
	go create_python_workflow(workflow_name, container_id_ch, error_ch)

	conatiner_id := <-container_id_ch
	c.String(http.StatusOK, conatiner_id)
}

func handle_ch_error(error_ch chan error, err error) {
	if err != nil {
		error_ch <- err
		panic(err)
	}
}

func create_python_workflow(workflow_name string, container_id_ch chan string, error_ch chan error) {
	output_ch := make(chan string, 1)
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
	handle_ch_error(error_ch, err)

	containers[resp.ID] = output_ch
	container_id_ch <- resp.ID
	close(container_id_ch)

	err = docker_cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	handle_ch_error(error_ch, err)

	// statusCh, errCh := docker_cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	// select {
	// case err := <-errCh:
	// 	handle_ch_error(error_ch, err)
	// case status := <-statusCh:
	// 	// Print the exit code of the container
	// 	fmt.Printf("Container exited with status %d\n", status.StatusCode)
	// }

	output_ch <- ">>>>>>> End"
	// return err
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

func clone_repo() {
	repoURL := "ssh://git@localhost:2220/srv/repo"
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

func pull_images() {
	fmt.Println("Pulling python:latest ...")
	ctx := context.Background()
	reader, err := docker_cli.ImagePull(ctx, "python:latest", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// Print the output from the pull
	buf := make([]byte, 4096)
	for {
		_, err := reader.Read(buf)
		if err != nil {
			break
		}
	}
	fmt.Println("Image pulled")
}

var error_group *errgroup.Group
var docker_cli *client.Client
var containers map[string]chan string

func main() {
	// docker_cli()
	// clone_repo()
	var err error
	containers = make(map[string]chan string)
	error_group = new(errgroup.Group)
	docker_cli, err = client.NewClientWithOpts(client.FromEnv)
	handle_error(err)
	pull_images()

	r := setup_router()
	r.Run(":8080")
}

func handle_error(err error) {
	if err != nil {
		panic(err)
	}
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
