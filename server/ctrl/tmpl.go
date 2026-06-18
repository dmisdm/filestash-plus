package ctrl

import "github.com/mickael-kerjean/filestash/server/common"

func TmplExec(params string, input map[string]string) (string, error) {
	return common.TmplExec(params, input)
}

func TmplParams(data map[string]string) map[string]string {
	return common.TmplParams(data)
}
