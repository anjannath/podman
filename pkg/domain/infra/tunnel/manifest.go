package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/containers/image/v5/types"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/bindings/manifests"
	"github.com/containers/podman/v4/pkg/domain/entities"
)

// ManifestCreate implements manifest create via ImageEngine
func (ir *ImageEngine) ManifestCreate(ctx context.Context, name string, images []string, opts entities.ManifestCreateOptions) (string, error) {
	options := new(manifests.CreateOptions).WithAll(opts.All).WithAmend(opts.Amend)
	imageID, err := manifests.Create(ir.ClientCtx, name, images, options)
	if err != nil {
		return imageID, fmt.Errorf("creating manifest: %w", err)
	}
	return imageID, err
}

// ManifestExists checks if a manifest list with the given name exists
func (ir *ImageEngine) ManifestExists(ctx context.Context, name string) (*entities.BoolReport, error) {
	exists, err := manifests.Exists(ir.ClientCtx, name, nil)
	if err != nil {
		return nil, err
	}
	return &entities.BoolReport{Value: exists}, nil
}

// ManifestInspect returns contents of manifest list with given name
func (ir *ImageEngine) ManifestInspect(_ context.Context, name string) ([]byte, error) {
	list, err := manifests.Inspect(ir.ClientCtx, name, nil)
	if err != nil {
		return nil, fmt.Errorf("getting content of manifest list or image %s: %w", name, err)
	}

	buf, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		return buf, fmt.Errorf("rendering manifest for display: %w", err)
	}
	return buf, err
}

// ManifestAdd adds images to the manifest list
func (ir *ImageEngine) ManifestAdd(_ context.Context, name string, imageNames []string, opts entities.ManifestAddOptions) (string, error) {
	options := new(manifests.AddOptions).WithAll(opts.All).WithAnnotation(opts.Annotation).WithArch(opts.Arch)
	options.WithVariant(opts.Variant).WithFeatures(opts.Features).WithImages(imageNames).WithOS(opts.OS)
	options.WithOSVersion(opts.OSVersion).WithUsername(opts.Username).WithPassword(opts.Password).WithAuthfile(opts.Authfile)
	if s := opts.SkipTLSVerify; s != types.OptionalBoolUndefined {
		if s == types.OptionalBoolTrue {
			options.WithSkipTLSVerify(true)
		} else {
			options.WithSkipTLSVerify(false)
		}
	}

	id, err := manifests.Add(ir.ClientCtx, name, options)
	if err != nil {
		return id, fmt.Errorf("adding to manifest list %s: %w", name, err)
	}
	return id, nil
}

// ManifestAnnotate updates an entry of the manifest list
func (ir *ImageEngine) ManifestAnnotate(ctx context.Context, name, images string, opts entities.ManifestAnnotateOptions) (string, error) {
	return "", errors.New("not implemented")
}

// ManifestRemoveDigest removes the digest from manifest list
func (ir *ImageEngine) ManifestRemoveDigest(ctx context.Context, name string, image string) (string, error) {
	updatedListID, err := manifests.Remove(ir.ClientCtx, name, image, nil)
	if err != nil {
		return updatedListID, fmt.Errorf("removing from manifest %s: %w", name, err)
	}
	return fmt.Sprintf("%s :%s\n", updatedListID, image), nil
}

// ManifestRm removes the specified manifest list from storage
func (ir *ImageEngine) ManifestRm(ctx context.Context, names []string) (*entities.ImageRemoveReport, []error) {
	return ir.Remove(ctx, names, entities.ImageRemoveOptions{LookupManifest: true})
}

// ManifestPush pushes a manifest list or image index to the destination
func (ir *ImageEngine) ManifestPush(ctx context.Context, name, destination string, opts entities.ImagePushOptions) (string, error) {
	options := new(images.PushOptions)
	options.WithUsername(opts.Username).WithPassword(opts.Password).WithAuthfile(opts.Authfile).WithRemoveSignatures(opts.RemoveSignatures).WithAll(opts.All).WithFormat(opts.Format).WithCompressionFormat(opts.CompressionFormat).WithQuiet(opts.Quiet).WithProgressWriter(opts.Writer)

	if s := opts.SkipTLSVerify; s != types.OptionalBoolUndefined {
		if s == types.OptionalBoolTrue {
			options.WithSkipTLSVerify(true)
		} else {
			options.WithSkipTLSVerify(false)
		}
	}
	digest, err := manifests.Push(ir.ClientCtx, name, destination, options)
	if err != nil {
		return "", fmt.Errorf("adding to manifest list %s: %w", name, err)
	}

	if opts.Rm {
		if _, rmErrors := ir.Remove(ctx, []string{name}, entities.ImageRemoveOptions{LookupManifest: true}); len(rmErrors) > 0 {
			return "", fmt.Errorf("removing manifest after push: %w", rmErrors[0])
		}
	}

	return digest, err
}
