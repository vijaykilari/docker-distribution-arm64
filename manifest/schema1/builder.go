package schema1

import (
	"fmt"

	"errors"
	"github.com/docker/distribution"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
	"github.com/docker/libtrust"
)

// ManifestBuilder is a type for constructing manifests
type manifestBuilder struct {
	Manifest
	pk libtrust.PrivateKey
}

// NewManifestBuilder is used to build new manifests for the current schema
// version.
func NewManifestBuilder(pk libtrust.PrivateKey, name, tag, architecture string) distribution.ManifestBuilder {
	return &manifestBuilder{
		Manifest: Manifest{
			Versioned: manifest.Versioned{
				SchemaVersion: 1,
			},
			Name:         name,
			Tag:          tag,
			Architecture: architecture,
		},
		pk: pk,
	}
}

// Build produces a final manifest from the given references
func (mb *manifestBuilder) Build() (distribution.Manifest, error) {
	m := mb.Manifest
	if len(m.FSLayers) == 0 {
		return nil, errors.New("cannot build manifest with zero layers or history")
	}

	m.FSLayers = make([]FSLayer, len(mb.Manifest.FSLayers))
	m.History = make([]History, len(mb.Manifest.History))
	copy(m.FSLayers, mb.Manifest.FSLayers)
	copy(m.History, mb.Manifest.History)

	return Sign(&m, mb.pk)
}

// AppendReference adds a reference to the current manifestBuilder
func (mb *manifestBuilder) AppendReference(d distribution.Describable) error {
	r, ok := d.(Reference)
	if !ok {
		return fmt.Errorf("Unable to add non-reference type to v1 builder")
	}

	// Entries need to be prepended
	mb.Manifest.FSLayers = append([]FSLayer{{BlobSum: r.Digest}}, mb.Manifest.FSLayers...)
	mb.Manifest.History = append([]History{r.History}, mb.Manifest.History...)
	return nil

}

// References returns the current references added to this builder
func (mb *manifestBuilder) References() []distribution.Descriptor {
	refs := make([]distribution.Descriptor, len(mb.Manifest.FSLayers))
	for i := range mb.Manifest.FSLayers {
		layerDigest := mb.Manifest.FSLayers[i].BlobSum
		history := mb.Manifest.History[i]
		ref := Reference{layerDigest, 0, history}
		refs[i] = ref.Descriptor()
	}
	return refs
}

// Reference describes a manifest v2, schema version 1 dependency.
// An FSLayer associated with a history entry.
type Reference struct {
	Digest  digest.Digest
	Size    int64 // if we know it, set it for the descriptor.
	History History
}

// Descriptor describes a reference
func (r Reference) Descriptor() distribution.Descriptor {
	return distribution.Descriptor{
		MediaType: MediaTypeManifestLayer,
		Digest:    r.Digest,
		Size:      r.Size,
	}
}
