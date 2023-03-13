package amazon

import (
	"fmt"
	"net/http"

	"webscrapers/internal"
)

//goland:noinspection GoUnusedExportedFunction
func Handler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	mobiAsinParamName := "mobiAsin"
	browserlessTokenParamName := "browserlessToken"

	if !queryParams.Has(mobiAsinParamName) || !queryParams.Has(browserlessTokenParamName) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(
			[]byte(fmt.Sprintf(
				"the query param %q and %q is required",
				mobiAsinParamName,
				browserlessTokenParamName,
			)),
		)
		return
	}

	asin := queryParams.Get(mobiAsinParamName)
	browserlessToken := queryParams.Get(browserlessTokenParamName)

	bookInfo, err := internal.GetBookInfo(browserlessToken, asin, "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Cache-Control", "max-age=0, s-maxage=86400")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(bookInfo.String()))
}
