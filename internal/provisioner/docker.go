package provisioner

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Aneesh-382005/Orbital/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	nat "github.com/docker/go-connections/nat"
)

const (
	codeServerImage = "codercom/code-server:latest"
	basePort        = 10000
	maxPort         = 11000
)

type DockerProvisioner struct {
	client *client.Client
}

func NewDockerProvisioner() (*DockerProvisioner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to docker client: %w", err)
	}
	return &DockerProvisioner{client: cli}, nil
}

func (p *DockerProvisioner) findFreePort(ctx context.Context) (int, error) {
	used := map[int]bool{}

	f := filters.NewArgs()
	f.Add("label",
		"orbital.managed=true")

	containers, err := p.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: f,
	})

	if err != nil {
		return 0, fmt.Errorf("listing containers: %w", err)
	}

	for _, c := range containers {
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				used[int(p.PublicPort)] = true
			}
		}
	}

	for port := basePort; port < maxPort; port++ {
		if !used[port] {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free ports available in range %d-%d", basePort, maxPort)
}
func (p *DockerProvisioner) StartWorkspace(ctx context.Context, ws *models.Workspace) error {
	// Pull image if not present
	reader, err := p.client.ImagePull(ctx, codeServerImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image: %w", err)
	}
	io.Copy(os.Stdout, reader)
	reader.Close()

	port, err := p.findFreePort(ctx)
	if err != nil {
		return err
	}

	containerPort := nat.Port("8080/tcp")
	hostPort := fmt.Sprintf("%d", port)

	resp, err := p.client.ContainerCreate(ctx,
		&container.Config{
			Image: codeServerImage,
			Env: []string{
				"PASSWORD=orbital-secret",
				fmt.Sprintf("WORKSPACE_ID=%s", ws.ID),
			},
			Labels: map[string]string{
				"orbital.managed":      "true",
				"orbital.workspace_id": ws.ID,
				"orbital.user_id":      ws.UserID,
			},
			ExposedPorts: nat.PortSet{containerPort: struct{}{}},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: hostPort},
				},
			},
			Resources: container.Resources{
				Memory:   512 * 1024 * 1024, // 512MB
				NanoCPUs: 1_000_000_000,     // 1 CPU
			},
			NetworkMode: "bridge",
		},
		&network.NetworkingConfig{},
		nil,
		fmt.Sprintf("orbital-%s", ws.ID),
	)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	if err := p.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	ws.ContainerID = resp.ID
	ws.Port = port
	ws.Status = models.StatusRunning
	ws.UpdatedAt = time.Now()

	return nil
}

func (p *DockerProvisioner) StopWorkspace(ctx context.Context, containerID string) error {
	timeout := 10
	if err := p.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stopping container: %w", err)
	}
	return nil
}

func (p *DockerProvisioner) RemoveWorkspace(ctx context.Context, containerID string) error {
	if err := p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("removing container: %w", err)
	}
	return nil
}
