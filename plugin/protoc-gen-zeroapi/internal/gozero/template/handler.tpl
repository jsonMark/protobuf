package {{ .Handler.PackageName }}

import (
	"net/http"
	"{{ .RootDir }}/internal/logic/{{ .Handler.PackageName }}"
	"{{ .RootDir }}/internal/svc"
    pb "{{ .Handler.InputOutputGoPath}}"

    "github.com/zeromicro/go-zero/rest/httpx"
)

func {{ .Handler.Name }}Handler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req pb.{{ .Handler.InputType }}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := {{ .Handler.PackageName }}.New{{ .Handler.Name }}Logic(r.Context(), svcCtx)
		resp, err := l.{{ .Handler.Name }}(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
