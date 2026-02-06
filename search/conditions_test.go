package search

import (
	"testing"
)

func TestParseSearchTypeFilters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		query         string
		expectedTypes []string
		expectedTerms []string
		caseSensitive bool
	}{
		{
			name:          "image type filter",
			query:         "vacation type:image",
			expectedTypes: []string{"image"},
			expectedTerms: []string{"vacation"},
		},
		{
			name:          "audio type filter",
			query:         "music type:audio",
			expectedTypes: []string{"audio"},
			expectedTerms: []string{"music"},
		},
		{
			name:          "music type filter (alias for audio)",
			query:         "song type:music",
			expectedTypes: []string{"audio"},
			expectedTerms: []string{"song"},
		},
		{
			name:          "video type filter",
			query:         "movie type:video",
			expectedTypes: []string{"video"},
			expectedTerms: []string{"movie"},
		},
		{
			name:          "pdf type filter (extension)",
			query:         "report type:pdf",
			expectedTypes: []string{"pdf"},
			expectedTerms: []string{"report"},
		},
		{
			name:          "jpg type filter (extension)",
			query:         "photo type:jpg",
			expectedTypes: []string{"jpg"},
			expectedTerms: []string{"photo"},
		},
		{
			name:          "multiple type filters",
			query:         "media type:image type:video",
			expectedTypes: []string{"image", "video"},
			expectedTerms: []string{"media"},
		},
		{
			name:          "type filter with case sensitivity",
			query:         "Document type:pdf case:sensitive",
			expectedTypes: []string{"pdf"},
			expectedTerms: []string{"Document"},
			caseSensitive: true,
		},
		{
			name:          "type filter with quoted term",
			query:         `"my photo" type:jpg`,
			expectedTypes: []string{"jpg"},
			expectedTerms: []string{"my photo"},
		},
		{
			name:          "only type filter",
			query:         "type:image",
			expectedTypes: []string{"image"},
			expectedTerms: []string{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts := parseSearch(tc.query)

			if len(opts.Conditions) != len(tc.expectedTypes) {
				t.Errorf("Expected %d conditions, got %d", len(tc.expectedTypes), len(opts.Conditions))
			}

			if opts.CaseSensitive != tc.caseSensitive {
				t.Errorf("Expected caseSensitive=%v, got %v", tc.caseSensitive, opts.CaseSensitive)
			}

			if len(opts.Terms) != len(tc.expectedTerms) {
				t.Errorf("Expected %d terms, got %d", len(tc.expectedTerms), len(opts.Terms))
			} else {
				for i, term := range tc.expectedTerms {
					if opts.Terms[i] != term {
						t.Errorf("Term %d: expected %q, got %q", i, term, opts.Terms[i])
					}
				}
			}
		})
	}
}

func TestConditionFunctions(t *testing.T) {
	t.Parallel()

	t.Run("imageCondition matches image files", func(t *testing.T) {
		t.Parallel()

		// Test with common image extensions
		imageTests := []struct {
			path     string
			expected bool
		}{
			{"/photos/image.jpg", true},
			{"/photos/image.png", true},
			{"/photos/image.gif", true},
			{"/photos/image.bmp", true},
			{"/photos/image.webp", true},
			{"/photos/image.tiff", true},
			{"/photos/document.pdf", false},
			{"/photos/audio.mp3", false},
			{"/photos/video.mp4", false},
			{"/photos/script.js", false},
			{"/photos/readme.txt", false},
		}

		for _, tt := range imageTests {
			result := imageCondition(tt.path)
			if result != tt.expected {
				t.Errorf("imageCondition(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		}
	})

	t.Run("audioCondition matches audio files", func(t *testing.T) {
		t.Parallel()

		audioTests := []struct {
			path     string
			expected bool
		}{
			{"/music/song.mp3", true},
			{"/music/song.wav", true},
			{"/music/song.ogg", true},
			{"/music/song.flac", true},
			{"/music/song.m4a", true},
			{"/music/video.mp4", false},
			{"/music/image.jpg", false},
			{"/music/document.pdf", false},
		}

		for _, tt := range audioTests {
			result := audioCondition(tt.path)
			if result != tt.expected {
				t.Errorf("audioCondition(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		}
	})

	t.Run("videoCondition matches video files", func(t *testing.T) {
		t.Parallel()

		videoTests := []struct {
			path     string
			expected bool
		}{
			{"/videos/movie.mp4", true},
			{"/videos/movie.avi", true},
			{"/videos/movie.mkv", true},
			{"/videos/movie.mov", true},
			{"/videos/movie.webm", true},
			{"/videos/audio.mp3", false},
			{"/videos/image.jpg", false},
		}

		for _, tt := range videoTests {
			result := videoCondition(tt.path)
			if result != tt.expected {
				t.Errorf("videoCondition(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		}
	})

	t.Run("extensionCondition matches specific extensions", func(t *testing.T) {
		t.Parallel()

		cond := extensionCondition("pdf")

		tests := []struct {
			path     string
			expected bool
		}{
			{"/documents/report.pdf", true},
			{"/documents/letter.PDF", false}, // case sensitive - filepath.Ext preserves case
			{"/documents/image.jpg", false},
			{"/documents/script.js", false},
		}

		for _, tt := range tests {
			result := cond(tt.path)
			if result != tt.expected {
				t.Errorf("extensionCondition('pdf')(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		}
	})
}
