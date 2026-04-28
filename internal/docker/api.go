package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// APIClient implements Client using Docker API
type APIClient struct {
	cli *client.Client
}

// PullImage pulls an image using Docker API, outputs progress to stdout
func (c *APIClient) PullImage(ctx context.Context, imageRef string) error {
	out, err := c.cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	decoder := json.NewDecoder(out)
	for {
		var msg struct {
			Status   string `json:"status"`
			ID       string `json:"id"`
			Progress string `json:"progress"`
			Error    string `json:"error"`
		}
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if msg.Error != "" {
			return fmt.Errorf("pull error: %s", msg.Error)
		}

		if msg.ID != "" {
			if msg.Progress != "" {
				fmt.Fprintf(os.Stdout, "%s: %s %s\n", msg.ID, msg.Status, msg.Progress)
			} else {
				fmt.Fprintf(os.Stdout, "%s: %s\n", msg.ID, msg.Status)
			}
		} else if msg.Status != "" {
			fmt.Fprintln(os.Stdout, msg.Status)
		}
	}

	fmt.Fprintln(os.Stdout, "Pull complete")
	return nil
}

// TagImage tags an image using Docker API
func (c *APIClient) TagImage(ctx context.Context, sourceRef, targetRef string) error {
	return c.cli.ImageTag(ctx, sourceRef, targetRef)
}

// GetImageSize returns image size using Docker API
func (c *APIClient) GetImageSize(ctx context.Context, imageRef string) (int64, error) {
	images, err := c.cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: imageRef,
		}),
	})
	if err != nil {
		return 0, err
	}
	if len(images) == 0 {
		return 0, fmt.Errorf("image not found: %s", imageRef)
	}
	return images[0].Size, nil
}