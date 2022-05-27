package images

import (
	"context"
	"strings"

	apiv1 "github.com/acorn-io/acorn/pkg/apis/api.acorn.io/v1"
	"github.com/acorn-io/acorn/pkg/pull"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewImageDetails(c client.WithWatch, images *Storage) *ImageDetails {
	return &ImageDetails{
		client: c,
		images: images,
	}
}

type ImageDetails struct {
	images *Storage
	client client.WithWatch
}

func (s *ImageDetails) NamespaceScoped() bool {
	return true
}

func (s *ImageDetails) New() runtime.Object {
	return &apiv1.ImageDetails{}
}

func (s *ImageDetails) Create(ctx context.Context, name string, obj runtime.Object, createValidation rest.ValidateObjectFunc, options *metav1.CreateOptions) (runtime.Object, error) {
	name = strings.ReplaceAll(name, "+", "/")

	if createValidation != nil {
		if err := createValidation(ctx, obj); err != nil {
			return nil, err
		}
	}

	ns, _ := request.NamespaceFrom(ctx)
	imageName := name

	image, err := s.images.ImageGet(ctx, name)
	if err != nil && !apierror.IsNotFound(err) {
		return nil, err
	} else if err == nil {
		ns = image.Namespace
		imageName = image.Name
	}

	appImage, err := pull.AppImage(ctx, s.client, ns, imageName)
	if err != nil {
		return nil, err
	}

	return &apiv1.ImageDetails{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageName,
			Namespace: ns,
		},
		AppImage: *appImage,
	}, nil
}
