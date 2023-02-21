package {{ .Logic.PackageName }}

import (
	"context"

	"{{ .RootDir }}/internal/svc"
	pb "{{ .Logic.InputOutputGoPath }}"

	"github.com/zeromicro/go-zero/core/logx"
)

type {{ .Logic.Name }}Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func New{{ .Logic.Name }}Logic(ctx context.Context, svcCtx *svc.ServiceContext) *{{ .Logic.Name }}Logic {
	return &{{ .Logic.Name }}Logic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *{{ .Logic.Name }}Logic) {{ .Logic.Name }}(req *pb.{{ .Logic.InputType }}) (*pb.{{ .Logic.OutputType }}, error) {
    // todo: add your logic here and delete this line

	return &pb.{{ .Logic.OutputType }}{}, nil
}
