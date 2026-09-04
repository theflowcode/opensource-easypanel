package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

type dockerBuildMessage struct {
	Stream string `json:"stream"`
	Error  string `json:"error"`
}

type dockerPullMessage struct {
	Status   string `json:"status"`
	Progress string `json:"progress"`
	ID       string `json:"id"`
	Error    string `json:"error"`
}

// BuildImage builds a Docker image from a local context directory and Dockerfile.
func (a *DockerAdapter) BuildImage(ctx context.Context, build domain.BuildConfig, logWriter io.Writer) (string, error) {
	tag := build.ImageTag
	if tag == "" {
		tag = fmt.Sprintf("openpanel/%s:latest", build.ServiceID)
	}

	dockerfile := build.DockerfilePath
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	tarBuffer, err := createTarArchive(build.ContextPath)
	if err != nil {
		return "", fmt.Errorf("failed to create tar archive from context %s: %w", build.ContextPath, err)
	}

	buildArgs := make(map[string]*string)
	for k, v := range build.BuildArgs {
		val := v
		buildArgs[k] = &val
	}

	opts := types.ImageBuildOptions{
		Tags:        []string{tag},
		Dockerfile:  dockerfile,
		BuildArgs:   buildArgs,
		Remove:      true,
		ForceRemove: true,
	}

	resp, err := a.cli.ImageBuild(ctx, tarBuffer, opts)
	if err != nil {
		return "", fmt.Errorf("image build call failed: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var msg dockerBuildMessage
		if err := decoder.Decode(&msg); err != nil {
			break
		}
		if msg.Error != "" {
			return "", fmt.Errorf("build error: %s", msg.Error)
		}
		if msg.Stream != "" && logWriter != nil {
			_, _ = logWriter.Write([]byte(msg.Stream))
		}
	}

	return tag, nil
}

// PullImage pulls an image from a container registry with optional credentials.
func (a *DockerAdapter) PullImage(ctx context.Context, imgName string, auth *domain.RegistryAuth, logWriter io.Writer) error {
	opts := image.PullOptions{}
	if auth != nil && auth.Username != "" {
		authConfig := registry.AuthConfig{
			Username:      auth.Username,
			Password:      auth.Password,
			ServerAddress: auth.ServerAddress,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("failed to encode registry auth: %w", err)
		}
		opts.RegistryAuth = base64.URLEncoding.EncodeToString(encodedJSON)
	}

	resp, err := a.cli.ImagePull(ctx, imgName, opts)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imgName, err)
	}
	defer resp.Close()

	decoder := json.NewDecoder(resp)
	for decoder.More() {
		var msg dockerPullMessage
		if err := decoder.Decode(&msg); err != nil {
			break
		}
		if msg.Error != "" {
			return fmt.Errorf("pull error: %s", msg.Error)
		}
		if logWriter != nil && msg.Status != "" {
			line := msg.Status
			if msg.ID != "" {
				line = fmt.Sprintf("[%s] %s", msg.ID, line)
			}
			if msg.Progress != "" {
				line = fmt.Sprintf("%s %s", line, msg.Progress)
			}
			_, _ = logWriter.Write([]byte(line + "\n"))
		}
	}

	return nil
}

// createTarArchive packs the context directory into a tar stream in memory.
func createTarArchive(contextPath string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	absContext, err := filepath.Abs(contextPath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(absContext, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(absContext, file)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			return nil
		}

		data, err := os.Open(file)
		if err != nil {
			return err
		}
		defer data.Close()

		_, err = io.Copy(tw, data)
		return err
	})

	if err != nil {
		return nil, err
	}
	return buf, nil
}
