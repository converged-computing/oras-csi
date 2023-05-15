package oras

import (
	"context"
	"io"
	"os"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
)

// Function that accepts an image reference and directory path
func pullBlob(target oras.ReadOnlyTarget, blobRef string, filename string) (fetchErr error) {

	ctx := context.Background()

	desc, err := target.Resolve(ctx, blobRef)
	if err != nil {
		return err
	}

	// Download using the descriptor
	readCloser, err := target.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	defer readCloser.Close()

	// Write the blob to a file
	log.Infof("OCI: Writing %s to %s", desc.Digest.String(), filename)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); fetchErr == nil {
			fetchErr = err
		}
	}()

	vr := content.NewVerifyReader(readCloser, desc)
	if _, err := io.Copy(file, vr); err != nil {
		return err
	}
	if err := vr.Verify(); err != nil {
		return err
	}

	return nil
}
