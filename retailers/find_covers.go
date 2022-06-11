package retailers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func FindCoverURL(artist string, title string) (string, error) {
	// use discogs
	//

	if strings.Index(title, ":") > 0 {
		title = title[0:strings.Index(title, ":")]
	}

	qA := url.QueryEscape(artist)
	qT := url.QueryEscape(title)
	query := fmt.Sprintf("https://www.discogs.com/search/?q=%s+%s+vinyl&type=all", qA, qT)

	resp, err := http.Get(query)
	if err != nil {
		return "", errors.Wrapf(err, "failed to retrieve search query %s", query)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "failed to extract body of search results %s", query)
	}
	image := ""
	toks := strings.Split(string(body), ">")
	for idx, t := range toks {
		if strings.Index(t, "thumbnail_center") >= 0 {
			image = toks[idx+1]
			image = image[strings.Index(image, "data-src=\"")+10:]
			image = image[0:strings.Index(image, "\"")]
			break
		}
	}
	return image, nil
}
