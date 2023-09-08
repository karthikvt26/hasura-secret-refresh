package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/hasura/hasura-secret-refresh/provider"
	"github.com/rs/zerolog"
)

func GetRequestRewriteDetails(
	rw http.ResponseWriter, r *http.Request, providers map[string]provider.Provider, requestLogger zerolog.Logger,
) (
	url *url.URL, headerKey string, headerVal string,
) {
	requestConfig, ok := getRequestConfig(rw, r, requestLogger)
	if !ok {
		return
	}
	url, ok = parseDestinationUrl(rw, r, requestConfig, requestLogger)
	if !ok {
		return
	}
	provider, ok := getProvider(rw, r, providers, requestConfig, requestLogger)
	if !ok {
		return
	}
	secret, ok := getSecret(rw, r, requestConfig, provider, requestLogger)
	if !ok {
		return
	}
	headerKey, headerVal, ok = getHeader(rw, r, secret, requestConfig, requestLogger)
	if !ok {
		return
	}
	return
}

func GetRequestRewriter(url *url.URL, headerKey string, headerVal string) func(req *httputil.ProxyRequest) {
	return func(req *httputil.ProxyRequest) {
		for k, _ := range RequestConfigs {
			req.Out.Header.Del(k)
		}
		req.SetURL(url)
		req.Out.Header.Set(headerKey, headerVal)
	}
}

func getRequestConfig(
	rw http.ResponseWriter, r *http.Request, requestLogger zerolog.Logger,
) (requestConfig RequestConfig, ok bool) {
	ok = true
	requestConfig, err := GetRequestConfig(r.Header)
	if err != nil {
		ok = false
		requestLogger.Error().Err(err).Msgf(err.Error())
		http.Error(rw, MakeHasuraError(err.Error()), http.StatusBadRequest)
		return
	}
	return
}

func parseDestinationUrl(
	rw http.ResponseWriter, r *http.Request,
	requestConfig RequestConfig, requestLogger zerolog.Logger,
) (url *url.URL, ok bool) {
	ok = true
	url, err := ParseUrl(requestConfig.DestinationUrl)
	if err != nil {
		ok = false
		requestLogger.Error().Msgf(err.Error())
		http.Error(rw, MakeHasuraError(err.Error()), http.StatusBadRequest)
		return
	}
	url = GetUrlWithSchemeAndHost(url)
	return
}

func getProvider(
	rw http.ResponseWriter, r *http.Request,
	providers map[string]provider.Provider, requestConfig RequestConfig, requestLogger zerolog.Logger,
) (provider provider.Provider, ok bool) {
	provider, ok = providers[requestConfig.SecretProvider]
	if !ok {
		errMsg := fmt.Sprintf("Provider name %s sent in header %s does not exist",
			requestConfig.SecretProvider, SecretProviderHeader)
		requestLogger.Error().Msgf(errMsg)
		http.Error(rw, MakeHasuraError(errMsg), http.StatusBadRequest)
		return
	}
	return
}

func getSecret(
	rw http.ResponseWriter, r *http.Request,
	requestConfig RequestConfig, provider provider.Provider, requestLogger zerolog.Logger,
) (secret string, ok bool) {
	ok = true
	secret, err := provider.GetSecret(requestConfig.SecretId)
	if err != nil {
		ok = false
		errMsg := fmt.Sprintf("Unable to fetch secret %s sent in header %s", requestConfig.SecretId, SecretIdHeader)
		requestLogger.Error().Err(err).Msgf(errMsg)
		http.Error(rw, MakeHasuraError(errMsg), http.StatusBadRequest)
		return
	}
	return
}

func getHeader(
	rw http.ResponseWriter, r *http.Request, secret string,
	requestConfig RequestConfig, requestLogger zerolog.Logger,
) (headerKey string, headerVal string, ok bool) {
	ok = true
	headerKey, headerVal, err := GetHeaderFromTemplate(requestConfig.HeaderTemplate, secret)
	if err != nil {
		ok = false
		errMsg := fmt.Sprintf("Header template %s sent in header %s is not valid", requestConfig.HeaderTemplate, TemplateHeader)
		requestLogger.Error().Err(err).Msgf(errMsg)
		http.Error(rw, MakeHasuraError(errMsg), http.StatusBadRequest)
		return
	}
	return
}
