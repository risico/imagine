package imagine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/risico/imagine"
)

func TestParamsFromQueryString(t *testing.T) {
	i, err := imagine.New(imagine.Params{
		Storage: imagine.NewInMemoryStorage(imagine.MemoryStoreParams{}),
		Cache:   imagine.NewInMemoryStorage(imagine.MemoryStoreParams{}),
	})
	assert.NoError(t, err)

	tests := []struct {
		name        string
		query       string
		expected    *imagine.ImageParams
		shouldError bool
	}{
		{
			name:  "basic width and height",
			query: "?w=300&h=200",
			expected: &imagine.ImageParams{
				Width:  300,
				Height: 200,
			},
		},
		{
			name:  "format conversion",
			query: "?format=webp",
			expected: &imagine.ImageParams{
				Format: "webp",
			},
		},
		{
			name:  "quality parameter",
			query: "?q=85",
			expected: &imagine.ImageParams{
				Quality: 85,
			},
		},
		{
			name:  "fit mode cover",
			query: "?w=300&h=200&fit=cover",
			expected: &imagine.ImageParams{
				Width:  300,
				Height: 200,
				Fit:    "cover",
			},
		},
		{
			name:  "rotation",
			query: "?rotate=90",
			expected: &imagine.ImageParams{
				Rotate: 90,
			},
		},
		{
			name:  "flip horizontal",
			query: "?flip=h",
			expected: &imagine.ImageParams{
				Flip: "h",
			},
		},
		{
			name:  "blur effect",
			query: "?blur=5.5",
			expected: &imagine.ImageParams{
				Blur: 5.5,
			},
		},
		{
			name:  "sharpen effect",
			query: "?sharpen=2",
			expected: &imagine.ImageParams{
				Sharpen: 2,
			},
		},
		{
			name:  "grayscale",
			query: "?grayscale",
			expected: &imagine.ImageParams{
				Grayscale: true,
			},
		},
		{
			name:  "gravity center",
			query: "?gravity=center",
			expected: &imagine.ImageParams{
				Gravity: "center",
			},
		},
		{
			name:  "thumbnail",
			query: "?thumbnail=150",
			expected: &imagine.ImageParams{
				Thumbnail: 150,
			},
		},
		{
			name:  "complex query",
			query: "?w=800&h=600&fit=cover&q=90&rotate=180&flip=both&gravity=smart&format=jpeg",
			expected: &imagine.ImageParams{
				Width:     800,
				Height:    600,
				Fit:       "cover",
				Quality:   90,
				Rotate:    180,
				Flip:      "both",
				Gravity:   "smart",
				Format:    "jpeg",
			},
		},
		{
			name:        "invalid quality too high",
			query:       "?q=101",
			shouldError: true,
		},
		{
			name:        "invalid quality too low",
			query:       "?q=0",
			shouldError: true,
		},
		{
			name:        "invalid fit mode",
			query:       "?fit=invalid",
			shouldError: true,
		},
		{
			name:        "invalid rotate angle",
			query:       "?rotate=45",
			shouldError: true,
		},
		{
			name:        "invalid flip mode",
			query:       "?flip=diagonal",
			shouldError: true,
		},
		{
			name:        "invalid blur too low",
			query:       "?blur=0.1",
			shouldError: true,
		},
		{
			name:        "invalid gravity",
			query:       "?gravity=invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := i.ParamsFromQueryString("http://example.com/image.jpg" + tt.query)
			
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, params)
			}
		})
	}
}