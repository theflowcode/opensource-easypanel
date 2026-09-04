package docker

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// StreamServiceLogs streams stdout and stderr logs from a service or its running container.
func (a *DockerAdapter) StreamServiceLogs(ctx context.Context, serviceID string, opts domain.LogStreamOptions, stdout, stderr io.Writer) error {
	tail := "all"
	if opts.TailLines > 0 {
		tail = strconv.Itoa(opts.TailLines)
	}

	containerIDs := a.getServiceContainerIDs(ctx, serviceID, serviceID)
	if len(containerIDs) > 0 {
		logReader, err := a.cli.ContainerLogs(ctx, containerIDs[0], container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     opts.Follow,
			Timestamps: opts.Timestamps,
			Tail:       tail,
		})
		if err != nil {
			return fmt.Errorf("failed to open container logs: %w", err)
		}
		defer logReader.Close()

		_, _ = stdcopy.StdCopy(stdout, stderr, logReader)
		return nil
	}

	svc, err := a.findService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("no running container or service found for %s: %w", serviceID, err)
	}

	logReader, err := a.cli.ServiceLogs(ctx, svc.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
		Tail:       tail,
	})
	if err != nil {
		return fmt.Errorf("failed to open service logs: %w", err)
	}
	defer logReader.Close()

	_, _ = stdcopy.StdCopy(stdout, stderr, logReader)
	return nil
}

// StreamDockerEvents listens to Docker daemon events and dispatches them to eventChan.
func (a *DockerAdapter) StreamDockerEvents(ctx context.Context, eventChan chan<- domain.DockerEvent) error {
	msgChan, errChan := a.cli.Events(ctx, events.ListOptions{})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case msg := <-msgChan:
			eventChan <- domain.DockerEvent{
				Type:   string(msg.Type),
				Action: string(msg.Action),
				Actor:  msg.Actor.ID,
				Time:   time.Unix(msg.Time, msg.TimeNano),
			}
		}
	}
}

// ExecServiceTerminal attaches an interactive pseudo-terminal (PTY) session to a container.
func (a *DockerAdapter) ExecServiceTerminal(ctx context.Context, serviceID string, stdin io.Reader, stdout, stderr io.Writer, resizeChan <-chan domain.TerminalSize) error {
	containerIDs := a.getServiceContainerIDs(ctx, serviceID, serviceID)
	if len(containerIDs) == 0 {
		return fmt.Errorf("no active container available for terminal in service: %s", serviceID)
	}
	containerID := containerIDs[0]

	execConfig := container.ExecOptions{
		AttachStdin:  stdin != nil,
		AttachStdout: stdout != nil,
		AttachStderr: stderr != nil,
		Tty:          true,
		Cmd:          []string{"/bin/sh"},
	}

	execResp, err := a.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create container exec instance: %w", err)
	}

	attachResp, err := a.cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		return fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer attachResp.Close()

	if resizeChan != nil {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case size, ok := <-resizeChan:
					if !ok {
						return
					}
					_ = a.cli.ContainerExecResize(ctx, execResp.ID, container.ResizeOptions{
						Height: uint(size.Rows),
						Width:  uint(size.Cols),
					})
				}
			}
		}()
	}

	errCopyChan := make(chan error, 1)
	if stdin != nil {
		go func() {
			_, _ = io.Copy(attachResp.Conn, stdin)
			_ = attachResp.CloseWrite()
		}()
	}

	go func() {
		_, err := io.Copy(stdout, attachResp.Reader)
		errCopyChan <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCopyChan:
		return err
	}
}
