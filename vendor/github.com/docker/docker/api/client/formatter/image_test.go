package formatter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/engine-api/types"
)

func TestImageContext(t *testing.T) {
	imageID := stringid.GenerateRandomID()
	unix := time.Now().Unix()

	var ctx imageContext
	cases := []struct {
		imageCtx  imageContext
		expValue  string
		expHeader string
		call      func() string
	}{
		{imageContext{
			i:     types.Image{ID: imageID},
			trunc: true,
		}, stringid.TruncateID(imageID), imageIDHeader, ctx.ID},
		{imageContext{
			i:     types.Image{ID: imageID},
			trunc: false,
		}, imageID, imageIDHeader, ctx.ID},
		{imageContext{
			i:     types.Image{Size: 10},
			trunc: true,
		}, "10 B", sizeHeader, ctx.Size},
		{imageContext{
			i:     types.Image{Created: unix},
			trunc: true,
		}, time.Unix(unix, 0).String(), createdAtHeader, ctx.CreatedAt},
		// FIXME
		// {imageContext{
		// 	i:     types.Image{Created: unix},
		// 	trunc: true,
		// }, units.HumanDuration(time.Unix(unix, 0)), createdSinceHeader, ctx.CreatedSince},
		{imageContext{
			i:    types.Image{},
			repo: "busybox",
		}, "busybox", repositoryHeader, ctx.Repository},
		{imageContext{
			i:   types.Image{},
			tag: "latest",
		}, "latest", tagHeader, ctx.Tag},
		{imageContext{
			i:      types.Image{},
			digest: "sha256:d149ab53f8718e987c3a3024bb8aa0e2caadf6c0328f1d9d850b2a2a67f2819a",
		}, "sha256:d149ab53f8718e987c3a3024bb8aa0e2caadf6c0328f1d9d850b2a2a67f2819a", digestHeader, ctx.Digest},
	}

	for _, c := range cases {
		ctx = c.imageCtx
		v := c.call()
		if strings.Contains(v, ",") {
			compareMultipleValues(t, v, c.expValue)
		} else if v != c.expValue {
			t.Fatalf("Expected %s, was %s\n", c.expValue, v)
		}

		h := ctx.fullHeader()
		if h != c.expHeader {
			t.Fatalf("Expected %s, was %s\n", c.expHeader, h)
		}
	}
}

func TestImageContextWrite(t *testing.T) {
	unixTime := time.Now().AddDate(0, 0, -1).Unix()
	expectedTime := time.Unix(unixTime, 0).String()

	contexts := []struct {
		context  ImageContext
		expected string
	}{
		// Errors
		{
			ImageContext{
				Context: Context{
					Format: "{{InvalidFunction}}",
				},
			},
			`Template parsing error: template: :1: function "InvalidFunction" not defined
`,
		},
		{
			ImageContext{
				Context: Context{
					Format: "{{nil}}",
				},
			},
			`Template parsing error: template: :1:2: executing "" at <nil>: nil is not a command
`,
		},
		// Table Format
		{
			ImageContext{
				Context: Context{
					Format: "table",
				},
			},
			`REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
image               tag1                imageID1            24 hours ago        0 B
image               tag2                imageID2            24 hours ago        0 B
<none>              <none>              imageID3            24 hours ago        0 B
`,
		},
		{
			ImageContext{
				Context: Context{
					Format: "table {{.Repository}}",
				},
			},
			"REPOSITORY\nimage\nimage\n<none>\n",
		},
		{
			ImageContext{
				Context: Context{
					Format: "table {{.Repository}}",
				},
				Digest: true,
			},
			`REPOSITORY          DIGEST
image               sha256:cbbf2f9a99b47fc460d422812b6a5adff7dfee951d8fa2e4a98caa0382cfbdbf
image               <none>
<none>              <none>
`,
		},
		{
			ImageContext{
				Context: Context{
					Format: "table {{.Repository}}",
					Quiet:  true,
				},
			},
			"REPOSITORY\nimage\nimage\n<none>\n",
		},
		{
			ImageContext{
				Context: Context{
					Format: "table",
					Quiet:  true,
				},
			},
			"imageID1\nimageID2\nimageID3\n",
		},
		{
			ImageContext{
				Context: Context{
					Format: "table",
					Quiet:  false,
				},
				Digest: true,
			},
			`REPOSITORY          TAG                 DIGEST                                                                    IMAGE ID            CREATED             SIZE
image               tag1                sha256:cbbf2f9a99b47fc460d422812b6a5adff7dfee951d8fa2e4a98caa0382cfbdbf   imageID1            24 hours ago        0 B
image               tag2                <none>                                                                    imageID2            24 hours ago        0 B
<none>              <none>              <none>                                                                    imageID3            24 hours ago        0 B
`,
		},
		{
			ImageContext{
				Context: Context{
					Format: "table",
					Quiet:  true,
				},
				Digest: true,
			},
			"imageID1\nimageID2\nimageID3\n",
		},
		// Raw Format
		{
			ImageContext{
				Context: Context{
					Format: "raw",
				},
			},
			fmt.Sprintf(`repository: image
tag: tag1
image_id: imageID1
created_at: %s
virtual_size: 0 B

repository: image
tag: tag2
image_id: imageID2
created_at: %s
virtual_size: 0 B

repository: <none>
tag: <none>
image_id: imageID3
created_at: %s
virtual_size: 0 B

`, expectedTime, expectedTime, expectedTime),
		},
		{
			ImageContext{
				Context: Context{
					Format: "raw",
				},
				Digest: true,
			},
			fmt.Sprintf(`repository: image
tag: tag1
digest: sha256:cbbf2f9a99b47fc460d422812b6a5adff7dfee951d8fa2e4a98caa0382cfbdbf
image_id: imageID1
created_at: %s
virtual_size: 0 B

repository: image
tag: tag2
digest: <none>
image_id: imageID2
created_at: %s
virtual_size: 0 B

repository: <none>
tag: <none>
digest: <none>
image_id: imageID3
created_at: %s
virtual_size: 0 B

`, expectedTime, expectedTime, expectedTime),
		},
		{
			ImageContext{
				Context: Context{
					Format: "raw",
					Quiet:  true,
				},
			},
			`image_id: imageID1
image_id: imageID2
image_id: imageID3
`,
		},
		// Custom Format
		{
			ImageContext{
				Context: Context{
					Format: "{{.Repository}}",
				},
			},
			"image\nimage\n<none>\n",
		},
		{
			ImageContext{
				Context: Context{
					Format: "{{.Repository}}",
				},
				Digest: true,
			},
			"image\nimage\n<none>\n",
		},
	}

	for _, context := range contexts {
		images := []types.Image{
			{ID: "imageID1", RepoTags: []string{"image:tag1"}, RepoDigests: []string{"image@sha256:cbbf2f9a99b47fc460d422812b6a5adff7dfee951d8fa2e4a98caa0382cfbdbf"}, Created: unixTime},
			{ID: "imageID2", RepoTags: []string{"image:tag2"}, Created: unixTime},
			{ID: "imageID3", RepoTags: []string{"<none>:<none>"}, RepoDigests: []string{"<none>@<none>"}, Created: unixTime},
		}
		out := bytes.NewBufferString("")
		context.context.Output = out
		context.context.Images = images
		context.context.Write()
		actual := out.String()
		if actual != context.expected {
			t.Fatalf("Expected \n%s, got \n%s", context.expected, actual)
		}
		// Clean buffer
		out.Reset()
	}
}

func TestImageContextWriteWithNoImage(t *testing.T) {
	out := bytes.NewBufferString("")
	images := []types.Image{}

	contexts := []struct {
		context  ImageContext
		expected string
	}{
		{
			ImageContext{
				Context: Context{
					Format: "{{.Repository}}",
					Output: out,
				},
			},
			"",
		},
		{
			ImageContext{
				Context: Context{
					Format: "table {{.Repository}}",
					Output: out,
				},
			},
			"REPOSITORY\n",
		},
		{
			ImageContext{
				Context: Context{
					Format: "{{.Repository}}",
					Output: out,
				},
				Digest: true,
			},
			"",
		},
		{
			ImageContext{
				Context: Context{
					Format: "table {{.Repository}}",
					Output: out,
				},
				Digest: true,
			},
			"REPOSITORY          DIGEST\n",
		},
	}

	for _, context := range contexts {
		context.context.Images = images
		context.context.Write()
		actual := out.String()
		if actual != context.expected {
			t.Fatalf("Expected \n%s, got \n%s", context.expected, actual)
		}
		// Clean buffer
		out.Reset()
	}
}
