package ceph

import cephdomain "cephtower/backend/internal/domain/ceph"

// Metadata is collected from rbd info in this pool's default namespace.
// Match both name and global ID: image recreation between reads must not join
// the previous image's status to the new image's configuration.
func enrichMirrorImageMetadata(value any, images []cephdomain.RBDImage) {
	rows, ok := value.([]any)
	if !ok {
		return
	}
	byName := make(map[string][]cephdomain.RBDImage)
	for _, image := range images {
		if image.Namespace == "" {
			byName[image.Name] = append(byName[image.Name], image)
		}
	}
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if !ok {
			continue
		}
		matches := byName[textField(row, "name")]
		if len(matches) != 1 || matches[0].MirrorGlobalID == "" || matches[0].MirrorGlobalID != textField(row, "global_id") {
			continue
		}
		image := matches[0]
		if image.MirrorMode == "journal" || image.MirrorMode == "snapshot" {
			row["mirror_mode"] = image.MirrorMode
		}
		// Disabled rbd info omits the global ID. Do not propagate the image
		// collector's default disabled state when the native state was absent.
		if image.MirrorState == "enabled" || image.MirrorState == "disabling" || image.MirrorState == "creating" {
			row["mirror_image_state"] = image.MirrorState
		}
		if image.Primary != nil {
			row["mirror_primary"] = *image.Primary
		}
	}
}
