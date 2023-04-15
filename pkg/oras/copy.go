package oras

import "oras.land/oras-go/v2"

// getCopyOptions returns the same copy options used for oras pull
func getCopyOptions(settings orasSettings) oras.CopyOptions {

	copyOptions := oras.DefaultCopyOptions
	copyOptions.Concurrency = settings.optionsConcurrency

	// This is taken from the oras client directly
	/*
		var getConfigOnce sync.Once
			copyOptions.FindSuccessors = func(ctx context.Context, fetcher content.Fetcher, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
				statusFetcher := content.FetcherFunc(func(ctx context.Context, target ocispec.Descriptor) (fetched io.ReadCloser, fetchErr error) {
					if _, ok := printed.LoadOrStore(generateContentKey(target), true); ok {
						return fetcher.Fetch(ctx, target)
					}

					// print status log for first-time fetching
					if err := display.PrintStatus(target, "Downloading", opts.Verbose); err != nil {
						return nil, err
					}
					rc, err := fetcher.Fetch(ctx, target)
					if err != nil {
						return nil, err
					}
					defer func() {
						if fetchErr != nil {
							rc.Close()
						}
					}()
					return rc, display.PrintStatus(target, "Processing ", opts.Verbose)
				})

				nodes, subject, config, err := graph.Successors(ctx, statusFetcher, desc)
				if err != nil {
					return nil, err
				}
				if subject != nil && opts.IncludeSubject {
					nodes = append(nodes, *subject)
				}
				if config != nil {
					getConfigOnce.Do(func() {
						if configPath != "" && (configMediaType == "" || config.MediaType == configMediaType) {
							if config.Annotations == nil {
								config.Annotations = make(map[string]string)
							}
							config.Annotations[ocispec.AnnotationTitle] = configPath
						}
					})
					nodes = append(nodes, *config)
				}

				var ret []ocispec.Descriptor
				for _, s := range nodes {
					if s.Annotations[ocispec.AnnotationTitle] == "" {
						ss, err := content.Successors(ctx, fetcher, s)
						if err != nil {
							return nil, err
						}
						if len(ss) == 0 {
							// skip s if it is unnamed AND has no successors.
							if err := printOnce(&printed, s, "Skipped    ", opts.Verbose); err != nil {
								return nil, err
							}
							continue
						}
					}
					ret = append(ret, s)
				}

				return ret, nil
			}

			target, err := opts.NewReadonlyTarget(ctx, opts.Common)
			if err != nil {
				return err
			}
			if err := opts.EnsureReferenceNotEmpty(); err != nil {
				return err
			}
			src, err := opts.CachedTarget(target)
			if err != nil {
				return err
			}
			dst, err := file.New(opts.Output)
			if err != nil {
				return err
			}
			defer dst.Close()
			dst.AllowPathTraversalOnWrite = opts.PathTraversal
			dst.DisableOverwrite = opts.KeepOldFiles

			pulledEmpty := true
			copyOptions.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
				if _, ok := printed.LoadOrStore(generateContentKey(desc), true); ok {
					return nil
				}
				return display.PrintStatus(desc, "Downloading", opts.Verbose)
			}
			copyOptions.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
				// restore named but deduplicated successor nodes
				successors, err := content.Successors(ctx, dst, desc)
				if err != nil {
					return err
				}
				for _, s := range successors {
					if _, ok := s.Annotations[ocispec.AnnotationTitle]; ok {
						if err := printOnce(&printed, s, "Restored   ", opts.Verbose); err != nil {
							return err
						}
					}
				}
				name, ok := desc.Annotations[ocispec.AnnotationTitle]
				if !ok {
					if !opts.Verbose {
						return nil
					}
					name = desc.MediaType
				} else {
					// named content downloaded
					pulledEmpty = false
				}
				printed.Store(generateContentKey(desc), true)
				return display.Print("Downloaded ", display.ShortDigest(desc), name)
			}

			// Copy
			desc, err := oras.Copy(ctx, src, opts.Reference, dst, opts.Reference, copyOptions)
			if err != nil {
				return err
			}
			if pulledEmpty {
				fmt.Println("Downloaded empty artifact")
			}
			fmt.Println("Pulled", opts.AnnotatedReference())
			fmt.Println("Digest:", desc.Digest)
			return nil
		(*/
	return copyOptions

}
