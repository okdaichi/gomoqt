package msf

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackaging_StringAndKnown(t *testing.T) {
	tests := map[string]struct {
		packaging Packaging
		expected  string
		known     bool
	}{
		"loc": {
			packaging: PackagingLOC,
			expected:  "loc",
			known:     true,
		},
		"media timeline": {
			packaging: PackagingMediaTimeline,
			expected:  "mediatimeline",
			known:     true,
		},
		"custom": {
			packaging: Packaging("custom"),
			expected:  "custom",
			known:     false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.packaging.String())
			assert.Equal(t, tt.known, tt.packaging.IsKnown())
		})
	}
}

func TestRole_StringAndKnown(t *testing.T) {
	tests := map[string]struct {
		role     Role
		expected string
		known    bool
	}{
		"video": {
			role:     RoleVideo,
			expected: "video",
			known:    true,
		},
		"audio": {
			role:     RoleAudio,
			expected: "audio",
			known:    true,
		},
		"custom": {
			role:     Role("custom"),
			expected: "custom",
			known:    false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.String())
			assert.Equal(t, tt.known, tt.role.IsKnown())
		})
	}
}

func TestCatalogValidate_Errors(t *testing.T) {
	tests := map[string]struct {
		catalog      Catalog
		errorMessage string
	}{
		"duplicate track identity": {
			catalog: Catalog{
				Version:          1,
				DefaultNamespace: "live/demo",
				Tracks: []Track{
					{Name: "video", Packaging: PackagingLOC, IsLive: new(true)},
					{Namespace: "live/demo", Name: "video", Packaging: PackagingLOC, IsLive: new(false)},
				},
			},
			errorMessage: "duplicate track identity",
		},
		"loc packaging requires isLive": {
			catalog: Catalog{
				Version: 1,
				Tracks:  []Track{{Name: "video", Packaging: PackagingLOC}},
			},
			errorMessage: "isLive is required for loc tracks",
		},
		"live track rejects duration": {
			catalog: Catalog{
				Version: 1,
				Tracks: []Track{{
					Name:          "video",
					Packaging:     PackagingLOC,
					IsLive:        new(true),
					TrackDuration: new(int64(1)),
				}},
			},
			errorMessage: "trackDuration must not be present when isLive is true",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := tt.catalog.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)
		})
	}
}

func TestCatalogValidate_RequiresVersionForIndependentCatalog(t *testing.T) {
	catalog := Catalog{
		Tracks: []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
	}

	err := catalog.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "catalog version is required")
}

func TestCatalogValidate_TimelineRequirements(t *testing.T) {
	catalog := Catalog{
		Version: 1,
		Tracks: []Track{
			{Name: "timeline", Packaging: PackagingMediaTimeline},
			{Name: "events", Packaging: PackagingEventTimeline, MimeType: "application/json", Depends: []string{"video"}},
		},
	}

	err := catalog.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mediatimeline tracks must use mimeType application/json")
	assert.Contains(t, err.Error(), "mediatimeline tracks must declare depends")
	assert.Contains(t, err.Error(), "eventType is required for eventtimeline tracks")
}

func TestCatalogRoundTrip_PreservesCustomFields(t *testing.T) {
	input := []byte(`{
		"version": 1,
		"generatedAt": 1234,
		"com.example.catalog": "premium",
		"tracks": [
			{
				"name": "video",
				"packaging": "loc",
				"isLive": true,
				"com.example.track": 7
			}
		]
	}`)

	var catalog Catalog
	require.NoError(t, json.Unmarshal(input, &catalog))
	require.Contains(t, catalog.ExtraFields, "com.example.catalog")
	require.Contains(t, catalog.Tracks[0].ExtraFields, "com.example.track")

	output, err := json.Marshal(catalog)
	require.NoError(t, err)
	assert.Contains(t, string(output), `"com.example.catalog":"premium"`)
	assert.Contains(t, string(output), `"com.example.track":7`)
}

func TestParseCatalogString_RoundTrip(t *testing.T) {
	input := `{
		"version": 1,
		"tracks": [
			{"name": "video", "packaging": "loc", "isLive": true}
		]
	}`

	catalog, err := ParseCatalogString(input)
	require.NoError(t, err)
	assert.Equal(t, 1, catalog.Version)
	require.Len(t, catalog.Tracks, 1)
	assert.Equal(t, "video", catalog.Tracks[0].Name)
	assert.Equal(t, PackagingLOC, catalog.Tracks[0].Packaging)
	require.NotNil(t, catalog.Tracks[0].IsLive)
	assert.True(t, *catalog.Tracks[0].IsLive)
}

func TestParseCatalog_RejectsDeltaJSON(t *testing.T) {
	_, err := ParseCatalog([]byte(`{"deltaUpdate": true, "addTracks": [{"name": "video", "packaging": "loc", "isLive": true}]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delta catalog fields are not allowed")
}

func TestParseCatalog_RejectsTrailingJSON(t *testing.T) {
	_, err := ParseCatalog([]byte(`{"version":1} {"extra":2}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after top-level value")
}

func TestParseCatalogDelta_RoundTrip(t *testing.T) {
	input := `{
		"deltaUpdate": true,
		"addTracks": [
			{"name": "video", "packaging": "loc", "isLive": true}
		]
	}`

	delta, err := ParseCatalogDeltaString(input)
	require.NoError(t, err)
	require.Len(t, delta.AddTracks, 1)
	assert.Equal(t, "video", delta.AddTracks[0].Name)
	assert.Equal(t, PackagingLOC, delta.AddTracks[0].Packaging)
	require.NotNil(t, delta.AddTracks[0].IsLive)
	assert.True(t, *delta.AddTracks[0].IsLive)
}

func TestParseCatalogDelta_RejectsIndependentJSON(t *testing.T) {
	_, err := ParseCatalogDelta([]byte(`{"version": 1, "tracks": [{"name": "video", "packaging": "loc", "isLive": true}]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "independent catalog fields are not allowed")
}

func TestParseCatalogDelta_RejectsTrailingJSON(t *testing.T) {
	_, err := ParseCatalogDelta([]byte(`{"deltaUpdate":true,"addTracks":[{"name":"video","packaging":"loc","isLive":true}]} {"extra":2}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "after top-level value")
}

func TestCatalogApplyDelta_PreservesDeclaredOperationOrder(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true), Bitrate: new(int64(1000))}},
	}

	var delta CatalogDelta
	require.NoError(t, json.Unmarshal([]byte(`{
		"deltaUpdate": true,
		"removeTracks": [{"name": "video"}],
		"addTracks": [{"name": "video", "packaging": "loc", "isLive": true, "bitrate": 2000}]
	}`), &delta))

	updated, err := base.ApplyDelta(delta)
	require.NoError(t, err)
	require.Len(t, updated.Tracks, 1)
	require.NotNil(t, updated.Tracks[0].Bitrate)
	assert.Equal(t, int64(2000), *updated.Tracks[0].Bitrate)
}

func TestCatalogApplyDelta_CloneTrackInheritsParent(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks: []Track{{
			Name:        "video-1080",
			Packaging:   PackagingLOC,
			IsLive:      new(true),
			Codec:       "av01",
			Width:       new(int64(1920)),
			Height:      new(int64(1080)),
			Bitrate:     new(int64(5000000)),
			RenderGroup: new(int64(1)),
		}},
	}
	delta := CatalogDelta{
		CloneTracks: []TrackClone{{
			Track: Track{
				Name:    "video-720",
				Width:   new(int64(1280)),
				Height:  new(int64(720)),
				Bitrate: new(int64(3000000)),
			},
			ParentName: "video-1080",
		}},
	}

	updated, err := base.ApplyDelta(delta)
	require.NoError(t, err)
	require.Len(t, updated.Tracks, 2)
	clone := updated.Tracks[1]
	assert.Equal(t, "video-720", clone.Name)
	assert.Equal(t, "av01", clone.Codec)
	assert.Equal(t, int64(1280), *clone.Width)
	assert.Equal(t, int64(720), *clone.Height)
	assert.Equal(t, int64(3000000), *clone.Bitrate)
	assert.Equal(t, int64(1), *clone.RenderGroup)
}

func TestCatalogApplyDelta_Errors(t *testing.T) {
	tests := map[string]struct {
		base         Catalog
		delta        CatalogDelta
		errorMessage string
	}{
		"remove unknown track": {
			base: Catalog{
				Version: 1,
				Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
			},
			delta: CatalogDelta{
				RemoveTracks: []TrackRef{{Name: "audio"}},
			},
			errorMessage: "cannot remove unknown track",
		},
		"clone unknown parent": {
			base: Catalog{
				Version: 1,
				Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
			},
			delta: CatalogDelta{
				CloneTracks: []TrackClone{{Track: Track{Name: "audio"}, ParentName: "missing"}},
			},
			errorMessage: "cannot clone unknown parent track",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := tt.base.ApplyDelta(tt.delta)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)
		})
	}
}

func TestCatalogApplyDelta_MergesMetadata(t *testing.T) {
	generatedAt := int64(1234)
	base := Catalog{
		Version:          1,
		DefaultNamespace: "live/demo",
		Tracks:           []Track{{Namespace: "live/demo", Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
		ExtraFields:      map[string]json.RawMessage{"base": json.RawMessage(`true`)},
	}
	delta := CatalogDelta{
		DefaultNamespace: "live/updated",
		GeneratedAt:      &generatedAt,
		IsComplete:       true,
		ExtraFields:      map[string]json.RawMessage{"delta": json.RawMessage(`{"ok":true}`)},
		AddTracks:        []Track{{Name: "audio", Packaging: PackagingLOC, IsLive: new(false)}},
	}

	updated, err := base.ApplyDelta(delta)
	require.NoError(t, err)
	assert.Equal(t, "live/updated", updated.DefaultNamespace)
	require.NotNil(t, updated.GeneratedAt)
	assert.Equal(t, generatedAt, *updated.GeneratedAt)
	assert.True(t, updated.IsComplete)
	assert.Contains(t, updated.ExtraFields, "base")
	assert.Contains(t, updated.ExtraFields, "delta")
	require.Len(t, updated.Tracks, 2)
	assert.Equal(t, "audio", updated.Tracks[1].Name)
}

func TestCatalogApplyDelta_RejectsChangingDefaultNamespaceForInheritedTracks(t *testing.T) {
	base := Catalog{
		Version:          1,
		DefaultNamespace: "live/demo",
		Tracks: []Track{{
			Name:      "video",
			Packaging: PackagingLOC,
			IsLive:    new(true),
		}},
	}
	delta := CatalogDelta{
		DefaultNamespace: "live/updated",
		AddTracks: []Track{{
			Namespace: "live/updated",
			Name:      "audio",
			Packaging: PackagingLOC,
			IsLive:    new(false),
		}},
	}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot change default namespace")
}

func TestCatalogDeltaValidate_RemoveTracksRejectExtraFields(t *testing.T) {
	var delta CatalogDelta
	require.NoError(t, json.Unmarshal([]byte(`{
		"deltaUpdate": true,
		"removeTracks": [{"name": "video", "codec": "av01"}]
	}`), &delta))

	err := delta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove track entries may contain only name and optional namespace")
}

func TestValidationError_EmptyProblems(t *testing.T) {
	err := &ValidationError{}
	assert.Equal(t, "msf: validation failed", err.Error())
}

func TestTrackID_String(t *testing.T) {
	tests := map[string]struct {
		id       TrackID
		expected string
	}{
		"with namespace": {
			id:       TrackID{Namespace: "live/demo", Name: "video"},
			expected: "live/demo/video",
		},
		"empty namespace": {
			id:       TrackID{Namespace: "", Name: "video"},
			expected: "video",
		},
		"inherited namespace sentinel": {
			id:       TrackID{Namespace: inheritedNamespaceSentinel, Name: "video"},
			expected: "video",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.String())
		})
	}
}

func TestTrackRoundTrip_AllFields(t *testing.T) {
	track := Track{
		Namespace:     "live/demo",
		Name:          "video",
		Packaging:     PackagingLOC,
		EventType:     "",
		Role:          RoleVideo,
		IsLive:        new(true),
		TargetLatency: new(int64(500)),
		Label:         "HD",
		RenderGroup:   new(int64(1)),
		AltGroup:      new(int64(2)),
		InitData:      "AAAA",
		Depends:       []string{"audio"},
		TemporalID:    new(int64(3)),
		SpatialID:     new(int64(1)),
		Codec:         "av01",
		MimeType:      "video/mp4",
		Framerate:     new(int64(30)),
		Timescale:     new(int64(90000)),
		Bitrate:       new(int64(5000000)),
		Width:         new(int64(1920)),
		Height:        new(int64(1080)),
		SampleRate:    new(int64(48000)),
		ChannelConfig: "stereo",
		DisplayWidth:  new(int64(1920)),
		DisplayHeight: new(int64(1080)),
		Language:      "en",
		TrackDuration: nil,
	}

	data, err := json.Marshal(track)
	require.NoError(t, err)

	var decoded Track
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, track.Namespace, decoded.Namespace)
	assert.Equal(t, track.Name, decoded.Name)
	assert.Equal(t, track.Packaging, decoded.Packaging)
	assert.Equal(t, track.Role, decoded.Role)
	require.NotNil(t, decoded.IsLive)
	assert.True(t, *decoded.IsLive)
	assert.Equal(t, int64(500), *decoded.TargetLatency)
	assert.Equal(t, "HD", decoded.Label)
	assert.Equal(t, int64(1), *decoded.RenderGroup)
	assert.Equal(t, int64(2), *decoded.AltGroup)
	assert.Equal(t, "AAAA", decoded.InitData)
	assert.Equal(t, []string{"audio"}, decoded.Depends)
	assert.Equal(t, int64(3), *decoded.TemporalID)
	assert.Equal(t, int64(1), *decoded.SpatialID)
	assert.Equal(t, "av01", decoded.Codec)
	assert.Equal(t, "video/mp4", decoded.MimeType)
	assert.Equal(t, int64(30), *decoded.Framerate)
	assert.Equal(t, int64(90000), *decoded.Timescale)
	assert.Equal(t, int64(5000000), *decoded.Bitrate)
	assert.Equal(t, int64(1920), *decoded.Width)
	assert.Equal(t, int64(1080), *decoded.Height)
	assert.Equal(t, int64(48000), *decoded.SampleRate)
	assert.Equal(t, "stereo", decoded.ChannelConfig)
	assert.Equal(t, int64(1920), *decoded.DisplayWidth)
	assert.Equal(t, int64(1080), *decoded.DisplayHeight)
	assert.Equal(t, "en", decoded.Language)
}

func TestTrackValidate_EventTimeline(t *testing.T) {
	tests := map[string]struct {
		track        Track
		errorMessage string
	}{
		"eventtimeline missing eventType": {
			track: Track{
				Name:      "events",
				Packaging: PackagingEventTimeline,
				MimeType:  "application/json",
				Depends:   []string{"video"},
			},
			errorMessage: "eventType is required for eventtimeline tracks",
		},
		"eventtimeline wrong mimeType": {
			track: Track{
				Name:      "events",
				Packaging: PackagingEventTimeline,
				EventType: "scene-change",
				MimeType:  "text/plain",
				Depends:   []string{"video"},
			},
			errorMessage: "eventtimeline tracks must use mimeType application/json",
		},
		"eventtimeline missing depends": {
			track: Track{
				Name:      "events",
				Packaging: PackagingEventTimeline,
				EventType: "scene-change",
				MimeType:  "application/json",
			},
			errorMessage: "eventtimeline tracks must declare depends",
		},
		"mediatimeline eventType must not be set": {
			track: Track{
				Name:      "timeline",
				Packaging: PackagingMediaTimeline,
				EventType: "invalid",
				MimeType:  "application/json",
				Depends:   []string{"video"},
			},
			errorMessage: "eventType must not be set for mediatimeline tracks",
		},
		"loc eventType must not be set": {
			track: Track{
				Name:      "video",
				Packaging: PackagingLOC,
				EventType: "invalid",
				IsLive:    new(true),
			},
			errorMessage: "eventType must not be set for loc tracks",
		},
		"other packaging eventType must not be set": {
			track: Track{
				Name:      "video",
				Packaging: PackagingCMAF,
				EventType: "invalid",
			},
			errorMessage: "eventType must only be set for eventtimeline tracks",
		},
		"track missing name": {
			track: Track{
				Packaging: PackagingLOC,
				IsLive:    new(true),
			},
			errorMessage: "name is required",
		},
		"track missing packaging": {
			track: Track{
				Name: "video",
			},
			errorMessage: "packaging is required",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tt.track.validate("test")
			require.NotEmpty(t, problems)
			assert.Contains(t, problems[0], tt.errorMessage)
		})
	}
}

func TestCatalogMarshalJSON_AllFields(t *testing.T) {
	generatedAt := int64(9999)
	catalog := Catalog{
		Version:     1,
		GeneratedAt: &generatedAt,
		IsComplete:  true,
		Tracks: []Track{{
			Name:      "video",
			Packaging: PackagingLOC,
			IsLive:    new(true),
		}},
	}

	data, err := json.Marshal(catalog)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"generatedAt":9999`)
	assert.Contains(t, string(data), `"isComplete":true`)
	assert.Contains(t, string(data), `"version":1`)
}

func TestCatalogUnmarshalJSON_IsComplete(t *testing.T) {
	input := `{"version":1,"isComplete":true}`
	catalog, err := ParseCatalogString(input)
	require.NoError(t, err)
	assert.True(t, catalog.IsComplete)
}

func TestTrack_applyOverrides_AllFields(t *testing.T) {
	base := Track{
		Name:      "video",
		Packaging: PackagingLOC,
		IsLive:    new(true),
	}

	override := Track{
		presentFields: map[string]struct{}{
			"namespace": {}, "name": {}, "packaging": {}, "eventType": {},
			"role": {}, "isLive": {}, "targetLatency": {}, "label": {},
			"renderGroup": {}, "altGroup": {}, "initData": {}, "depends": {},
			"temporalId": {}, "spatialId": {}, "codec": {}, "mimeType": {},
			"framerate": {}, "timescale": {}, "bitrate": {}, "width": {},
			"height": {}, "samplerate": {}, "channelConfig": {}, "displayWidth": {},
			"displayHeight": {}, "lang": {}, "trackDuration": {},
		},
		Namespace:     "live/new",
		Name:          "audio",
		Packaging:     PackagingCMAF,
		EventType:     "scene",
		Role:          RoleAudio,
		IsLive:        new(false),
		TargetLatency: new(int64(200)),
		Label:         "SD",
		RenderGroup:   new(int64(5)),
		AltGroup:      new(int64(3)),
		InitData:      "BBBB",
		Depends:       []string{"sub"},
		TemporalID:    new(int64(2)),
		SpatialID:     new(int64(0)),
		Codec:         "opus",
		MimeType:      "audio/opus",
		Framerate:     new(int64(60)),
		Timescale:     new(int64(48000)),
		Bitrate:       new(int64(128000)),
		Width:         new(int64(640)),
		Height:        new(int64(480)),
		SampleRate:    new(int64(44100)),
		ChannelConfig: "mono",
		DisplayWidth:  new(int64(640)),
		DisplayHeight: new(int64(480)),
		Language:      "ja",
		TrackDuration: new(int64(60000)),
	}

	base.applyOverrides(override)

	assert.Equal(t, "live/new", base.Namespace)
	assert.Equal(t, "audio", base.Name)
	assert.Equal(t, PackagingCMAF, base.Packaging)
	assert.Equal(t, "scene", base.EventType)
	assert.Equal(t, RoleAudio, base.Role)
	require.NotNil(t, base.IsLive)
	assert.False(t, *base.IsLive)
	assert.Equal(t, int64(200), *base.TargetLatency)
	assert.Equal(t, "SD", base.Label)
	assert.Equal(t, int64(5), *base.RenderGroup)
	assert.Equal(t, int64(3), *base.AltGroup)
	assert.Equal(t, "BBBB", base.InitData)
	assert.Equal(t, []string{"sub"}, base.Depends)
	assert.Equal(t, int64(2), *base.TemporalID)
	assert.Equal(t, int64(0), *base.SpatialID)
	assert.Equal(t, "opus", base.Codec)
	assert.Equal(t, "audio/opus", base.MimeType)
	assert.Equal(t, int64(60), *base.Framerate)
	assert.Equal(t, int64(48000), *base.Timescale)
	assert.Equal(t, int64(128000), *base.Bitrate)
	assert.Equal(t, int64(640), *base.Width)
	assert.Equal(t, int64(480), *base.Height)
	assert.Equal(t, int64(44100), *base.SampleRate)
	assert.Equal(t, "mono", base.ChannelConfig)
	assert.Equal(t, int64(640), *base.DisplayWidth)
	assert.Equal(t, int64(480), *base.DisplayHeight)
	assert.Equal(t, "ja", base.Language)
	assert.Equal(t, int64(60000), *base.TrackDuration)
}

func TestCatalogApplyDelta_AddDuplicateTrack(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
	}
	delta := CatalogDelta{
		AddTracks: []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(false)}},
	}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add duplicate track")
}

func TestCatalogApplyDelta_CloneDuplicateTrack(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks: []Track{
			{Name: "video", Packaging: PackagingLOC, IsLive: new(true)},
			{Name: "video-copy", Packaging: PackagingLOC, IsLive: new(true)},
		},
	}
	delta := CatalogDelta{
		CloneTracks: []TrackClone{{
			Track:      Track{Name: "video-copy"},
			ParentName: "video",
		}},
	}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot clone into duplicate track")
}

func TestCatalogApplyDelta_CloneMissingName(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
	}
	delta := CatalogDelta{
		CloneTracks: []TrackClone{{
			Track:      Track{},
			ParentName: "video",
		}},
	}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestCatalogApplyDelta_InvalidBaseCatalog(t *testing.T) {
	base := Catalog{}
	delta := CatalogDelta{
		AddTracks: []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
	}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "catalog version is required")
}

func TestCatalogApplyDelta_InvalidDelta(t *testing.T) {
	base := Catalog{
		Version: 1,
		Tracks:  []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
	}
	delta := CatalogDelta{}

	_, err := base.ApplyDelta(delta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delta catalog must contain")
}

func TestDecodeOrderedObject_InvalidJSON(t *testing.T) {
	tests := map[string]struct {
		input        string
		errorMessage string
	}{
		"not an object": {
			input:        `[1,2,3]`,
			errorMessage: "expected JSON object",
		},
		"empty input": {
			input:        ``,
			errorMessage: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := decodeOrderedObject([]byte(tt.input))
			require.Error(t, err)
			if tt.errorMessage != "" {
				assert.Contains(t, err.Error(), tt.errorMessage)
			}
		})
	}
}

func TestPackaging_IsKnown_AllVariants(t *testing.T) {
	tests := map[string]struct {
		packaging Packaging
		known     bool
	}{
		"event timeline": {
			packaging: PackagingEventTimeline,
			known:     true,
		},
		"cmaf": {
			packaging: PackagingCMAF,
			known:     true,
		},
		"legacy": {
			packaging: PackagingLegacy,
			known:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.known, tt.packaging.IsKnown())
		})
	}
}

func TestRole_IsKnown_AllVariants(t *testing.T) {
	tests := map[string]struct {
		role  Role
		known bool
	}{
		"audiodescription": {
			role:  RoleAudioDescription,
			known: true,
		},
		"caption": {
			role:  RoleCaption,
			known: true,
		},
		"subtitle": {
			role:  RoleSubtitle,
			known: true,
		},
		"signlanguage": {
			role:  RoleSignLanguage,
			known: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.known, tt.role.IsKnown())
		})
	}
}

func TestCatalogMarshalJSON_EmptyCatalog(t *testing.T) {
	catalog := Catalog{}
	data, err := json.Marshal(catalog)
	require.NoError(t, err)
	assert.Equal(t, `{}`, string(data))
}

func TestTrack_hasField_UnknownField(t *testing.T) {
	track := Track{}
	assert.False(t, track.hasField("unknownField"))
}

func TestTrack_hasField_MateriallySet(t *testing.T) {
	track := Track{
		Namespace:     "live",
		Name:          "video",
		Packaging:     PackagingLOC,
		EventType:     "test",
		Role:          RoleVideo,
		IsLive:        new(true),
		TargetLatency: new(int64(100)),
		Label:         "HD",
		RenderGroup:   new(int64(1)),
		AltGroup:      new(int64(1)),
		InitData:      "AA",
		Depends:       []string{"a"},
		TemporalID:    new(int64(0)),
		SpatialID:     new(int64(0)),
		Codec:         "av01",
		MimeType:      "video/mp4",
		Framerate:     new(int64(30)),
		Timescale:     new(int64(90000)),
		Bitrate:       new(int64(5000)),
		Width:         new(int64(1920)),
		Height:        new(int64(1080)),
		SampleRate:    new(int64(48000)),
		ChannelConfig: "stereo",
		DisplayWidth:  new(int64(1920)),
		DisplayHeight: new(int64(1080)),
		Language:      "en",
		TrackDuration: new(int64(60000)),
	}

	fields := []string{
		"namespace", "name", "packaging", "eventType", "role", "isLive",
		"targetLatency", "label", "renderGroup", "altGroup", "initData",
		"depends", "temporalId", "spatialId", "codec", "mimeType",
		"framerate", "timescale", "bitrate", "width", "height",
		"samplerate", "channelConfig", "displayWidth", "displayHeight",
		"lang", "trackDuration",
	}

	for _, field := range fields {
		assert.True(t, track.hasField(field), "hasField(%q) should be true", field)
	}
}

func TestTrackDuration_VODRoundTrip(t *testing.T) {
	input := `{
		"version": 1,
		"tracks": [{
			"name": "video",
			"packaging": "loc",
			"isLive": false,
			"trackDuration": 120000
		}]
	}`

	catalog, err := ParseCatalogString(input)
	require.NoError(t, err)
	require.Len(t, catalog.Tracks, 1)
	require.NotNil(t, catalog.Tracks[0].TrackDuration)
	assert.Equal(t, int64(120000), *catalog.Tracks[0].TrackDuration)

	data, err := json.Marshal(catalog)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"trackDuration":120000`)
}
func TestCatalogDelta_Clone(t *testing.T) {
	generatedAt := int64(5000)
	delta := CatalogDelta{
		DefaultNamespace: "live/demo",
		GeneratedAt:      &generatedAt,
		IsComplete:       true,
		AddTracks:        []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
		RemoveTracks:     []TrackRef{{Name: "old", Namespace: "ns"}},
		CloneTracks:      []TrackClone{{Track: Track{Name: "video-720"}, ParentName: "video-1080"}},
		ExtraFields:      map[string]json.RawMessage{"ext": json.RawMessage(`1`)},
	}

	clone := delta.Clone()
	assert.Equal(t, delta.DefaultNamespace, clone.DefaultNamespace)
	require.NotNil(t, clone.GeneratedAt)
	assert.Equal(t, generatedAt, *clone.GeneratedAt)
	assert.True(t, clone.IsComplete)
	require.Len(t, clone.AddTracks, 1)
	assert.Equal(t, "video", clone.AddTracks[0].Name)
	require.Len(t, clone.RemoveTracks, 1)
	assert.Equal(t, "old", clone.RemoveTracks[0].Name)
	require.Len(t, clone.CloneTracks, 1)
	assert.Equal(t, "video-720", clone.CloneTracks[0].Name)
	assert.Contains(t, clone.ExtraFields, "ext")

	// Mutating clone should not affect original.
	clone.AddTracks[0].Name = "mutated"
	assert.Equal(t, "video", delta.AddTracks[0].Name)
}

func TestCatalogDelta_MarshalJSON_RoundTrip(t *testing.T) {
	generatedAt := int64(42)
	delta := CatalogDelta{
		GeneratedAt: &generatedAt,
		IsComplete:  true,
		AddTracks:   []Track{{Name: "video", Packaging: PackagingLOC, IsLive: new(true)}},
		RemoveTracks: []TrackRef{{
			Namespace: "live",
			Name:      "old",
		}},
		CloneTracks: []TrackClone{{
			Track:      Track{Name: "video-720", Width: new(int64(1280))},
			ParentName: "video-1080",
		}},
		ExtraFields: map[string]json.RawMessage{"custom": json.RawMessage(`true`)},
	}

	data, err := json.Marshal(delta)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"deltaUpdate":true`)
	assert.Contains(t, string(data), `"generatedAt":42`)
	assert.Contains(t, string(data), `"isComplete":true`)
	assert.Contains(t, string(data), `"addTracks"`)
	assert.Contains(t, string(data), `"removeTracks"`)
	assert.Contains(t, string(data), `"cloneTracks"`)
	assert.Contains(t, string(data), `"custom":true`)

	var decoded CatalogDelta
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.NotNil(t, decoded.GeneratedAt)
	assert.Equal(t, int64(42), *decoded.GeneratedAt)
	assert.True(t, decoded.IsComplete)
	require.Len(t, decoded.AddTracks, 1)
	assert.Equal(t, "video", decoded.AddTracks[0].Name)
	require.Len(t, decoded.RemoveTracks, 1)
	assert.Equal(t, "old", decoded.RemoveTracks[0].Name)
	require.Len(t, decoded.CloneTracks, 1)
	assert.Equal(t, "video-720", decoded.CloneTracks[0].Name)
	assert.Equal(t, "video-1080", decoded.CloneTracks[0].ParentName)
}

func TestTrackRef_MarshalJSON_RoundTrip(t *testing.T) {
	ref := TrackRef{
		Namespace: "live/demo",
		Name:      "video",
	}

	data, err := json.Marshal(ref)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"namespace":"live/demo"`)
	assert.Contains(t, string(data), `"name":"video"`)

	var decoded TrackRef
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "live/demo", decoded.Namespace)
	assert.Equal(t, "video", decoded.Name)
}

func TestTrackRef_Clone(t *testing.T) {
	ref := TrackRef{
		Namespace:   "live",
		Name:        "video",
		ExtraFields: map[string]json.RawMessage{"x": json.RawMessage(`1`)},
	}

	clone := ref.Clone()
	assert.Equal(t, ref.Namespace, clone.Namespace)
	assert.Equal(t, ref.Name, clone.Name)
	assert.Contains(t, clone.ExtraFields, "x")

	clone.ExtraFields["x"] = json.RawMessage(`2`)
	assert.Equal(t, json.RawMessage(`1`), ref.ExtraFields["x"])
}

func TestTrackRef_effectiveNamespace(t *testing.T) {
	tests := map[string]struct {
		ref              TrackRef
		defaultNamespace string
		expected         string
	}{
		"explicit namespace": {
			ref:              TrackRef{Namespace: "live"},
			defaultNamespace: "other",
			expected:         "live",
		},
		"inherits default": {
			ref:              TrackRef{},
			defaultNamespace: "live/demo",
			expected:         "live/demo",
		},
		"sentinel when both empty": {
			ref:              TrackRef{},
			defaultNamespace: "",
			expected:         inheritedNamespaceSentinel,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ref.effectiveNamespace(tt.defaultNamespace))
		})
	}
}

func TestTrackClone_MarshalJSON_RoundTrip(t *testing.T) {
	clone := TrackClone{
		Track: Track{
			Name:    "video-720",
			Width:   new(int64(1280)),
			Height:  new(int64(720)),
			Bitrate: new(int64(3000000)),
		},
		ParentName: "video-1080",
	}

	data, err := json.Marshal(clone)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"parentName":"video-1080"`)
	assert.Contains(t, string(data), `"name":"video-720"`)

	var decoded TrackClone
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "video-720", decoded.Name)
	assert.Equal(t, "video-1080", decoded.ParentName)
	require.NotNil(t, decoded.Width)
	assert.Equal(t, int64(1280), *decoded.Width)
}

func TestTrackClone_Clone(t *testing.T) {
	original := TrackClone{
		Track:      Track{Name: "video-720", Codec: "av01"},
		ParentName: "video-1080",
	}

	clone := original.Clone()
	assert.Equal(t, "video-720", clone.Name)
	assert.Equal(t, "video-1080", clone.ParentName)
	assert.Equal(t, "av01", clone.Codec)

	clone.Name = "mutated"
	assert.Equal(t, "video-720", original.Name)
}

func TestTrackClone_Validate(t *testing.T) {
	tests := map[string]struct {
		clone        TrackClone
		errorMessage string
	}{
		"missing name": {
			clone:        TrackClone{Track: Track{}, ParentName: "parent"},
			errorMessage: "name is required",
		},
		"missing parentName": {
			clone:        TrackClone{Track: Track{Name: "video"}},
			errorMessage: "parentName is required for clone tracks",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			problems := tt.clone.Validate("test")
			require.NotEmpty(t, problems)
			assert.Contains(t, problems[0], tt.errorMessage)
		})
	}
}

func TestCatalogDeltaValidate_EmptyDelta(t *testing.T) {
	delta := CatalogDelta{}
	err := delta.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delta catalog must contain")
}

func TestCatalogDelta_UnmarshalJSON_MissingDeltaUpdate(t *testing.T) {
	_, err := ParseCatalogDelta([]byte(`{"addTracks":[{"name":"video","packaging":"loc","isLive":true}]}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deltaUpdate=true")
}

func TestCatalogDelta_UnmarshalJSON_DeltaUpdateFalse(t *testing.T) {
	_, err := ParseCatalogDelta([]byte(`{"deltaUpdate":false}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deltaUpdate=true")
}

func TestCatalogDelta_UnmarshalJSON_CloneTracks(t *testing.T) {
	input := `{
		"deltaUpdate": true,
		"cloneTracks": [{"name": "video-720", "parentName": "video-1080", "width": 1280}]
	}`

	delta, err := ParseCatalogDeltaString(input)
	require.NoError(t, err)
	require.Len(t, delta.CloneTracks, 1)
	assert.Equal(t, "video-720", delta.CloneTracks[0].Name)
	assert.Equal(t, "video-1080", delta.CloneTracks[0].ParentName)
	require.NotNil(t, delta.CloneTracks[0].Width)
	assert.Equal(t, int64(1280), *delta.CloneTracks[0].Width)
}

func TestCatalogDelta_UnmarshalJSON_ExtraFields(t *testing.T) {
	input := `{
		"deltaUpdate": true,
		"addTracks": [{"name": "video", "packaging": "loc", "isLive": true}],
		"com.example.ext": 42
	}`

	delta, err := ParseCatalogDeltaString(input)
	require.NoError(t, err)
	assert.Contains(t, delta.ExtraFields, "com.example.ext")
}

func TestTrackRef_MarshalJSON_NameOnly(t *testing.T) {
	ref := TrackRef{Name: "video"}

	data, err := json.Marshal(ref)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "video", decoded["name"])
	_, hasNS := decoded["namespace"]
	assert.False(t, hasNS)
}
