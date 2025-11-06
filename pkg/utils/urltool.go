package utils

import (
	"encoding/json"
	"github/heimaolst/collectionbox/internal/biz"
	"log"
	"net/url"
	"os"
	"strings"
)

type Origins struct {
	Supports []string
	Items    []Item
}
type Item struct {
	Host   string
	Origin string
}

var OriginMap map[string]string = map[string]string{}

func init() {
	datas, err := os.ReadFile("resource/origin.json")
	if err != nil {
		log.Fatal("the origin.json can't read")
	}
	var origins Origins
	if err := json.Unmarshal(datas, &origins); err != nil {
		log.Fatal("the origin.json can't be unmarshaled")
	}
	for _, v := range origins.Items {
		OriginMap[v.Host] = v.Origin
	}
}
func GetOrgin(str string) (string, error) {
	url, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	host := strings.TrimPrefix(url.Host, "www.")

	if _, ok := OriginMap[host]; !ok {
		return "", biz.ErrInternalError.WithMessage("undefined origin")
	}
	return OriginMap[host], nil
}
