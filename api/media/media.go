package media

import (
	"context"
	"fmt"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"net/http"
)

var mediaFs afero.Fs

func WriteAvatar(data []byte, group string, fileUuid uuid.UUID) (err error) {
	fileName := fmt.Sprintf("avatars/%s/%s.png", group, fileUuid.String())
	err = afero.WriteFile(mediaFs, fileName, data, 0644)
	return
}

func GetAvatarUrl(group string, fileUuid uuid.UUID) string {
	return fmt.Sprintf("avatars/%s/%s.png", group, fileUuid.String())
}

func SetupLocal(api huma.API) {
	mediaFs = afero.NewBasePathFs(afero.NewOsFs(), ".tempmedia")

	mediaApi := huma.NewGroup(api, "/media")
	huma.Register(mediaApi, huma.Operation{
		Path:        "/avatars/{group}/{avatar}",
		OperationID: "media-avatars",
		Method:      http.MethodGet,
		Errors: []int{
			http.StatusNotFound,
		},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The contents will be an image",
				Content: map[string]*huma.MediaType{
					"image/png": {},
				},
			},
		},
		Tags:        []string{"Local/Avatars"},
		Metadata:    map[string]any{"NoUserAuth": true},
		Summary:     "Get an avatar image",
		Description: "Retrieves an avatar image. This endpoint only exists when running in a local environment, and is not intended to be invoked explicitly. You should always use the URL returned by other responses such as `User.AvatarUrl`.",
	}, getAvatar)
}

type AvatarInput struct {
	Group  string `path:"group"`
	Avatar string `path:"avatar"`
}

type AvatarOutput struct {
	Body []byte
}

func getAvatar(_ context.Context, input *AvatarInput) (output *AvatarOutput, err error) {
	filePath := fmt.Sprintf("avatars/%s/%s.png", input.Group, input.Avatar)

	var fileBytes []byte
	// TODO: shouldn't this require a context?
	fileBytes, err = afero.ReadFile(mediaFs, filePath)
	if err != nil {
		return
	}

	return &AvatarOutput{Body: fileBytes}, nil
}
