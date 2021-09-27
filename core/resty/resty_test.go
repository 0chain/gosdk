package resty

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/0chain/gosdk/core/resty/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResty(t *testing.T) {

	tests := []struct {
		name string
		urls []string

		statusCode  int
		exceptedErr error

		setup func(a *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty)
	}{
		{
			name:        "Test_Resty_Timeout",
			exceptedErr: context.DeadlineExceeded,
			urls:        []string{"Test_Resty_Timeout_1", "Test_Resty_Timeout_2", "Test_Resty_Timeout_3"},
			setup: func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {

				r := New(&http.Transport{}, nil)
				r.client = &mocks.Timeout{
					Timeout: 1 * time.Second,
				}

				ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
				go func() {
					<-ctx.Done()
					cancel()
				}()

				return ctx, r
			},
		},
		{
			name:        "Test_Resty_All_Success",
			statusCode:  200,
			exceptedErr: context.DeadlineExceeded,
			urls:        []string{"http://Test_Resty_Timeout_1", "http://Test_Resty_Timeout_2"},
			setup: func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {

				resty := New(&http.Transport{}, func(req *http.Request, resp *http.Response, cf context.CancelFunc, e error) error {

					ra.Equal(200, resp.StatusCode)

					buf, err := ioutil.ReadAll(resp.Body)
					defer resp.Body.Close()

					ra.Equal(nil, err)
					ra.Equal(name, string(buf))

					return nil
				})

				client := &mocks.Client{}

				for _, url := range urls {

					func(u string) {
						client.On("Do", mock.MatchedBy(func(r *http.Request) bool {
							return r.URL.String() == u
						})).Return(&http.Response{
							StatusCode: statusCode,
							Body:       ioutil.NopCloser(strings.NewReader(name)),
						}, nil)
					}(url)

				}

				resty.client = client

				ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
				go func() {
					<-ctx.Done()
					cancel()
				}()

				return context.TODO(), resty
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := require.New(t)

			ctx, resty := tt.setup(r, tt.name, tt.statusCode, tt.urls)

			resty.DoGet(ctx, tt.urls...)

			errs := resty.Wait()

			for _, err := range errs {
				// test it by predefined error variable instead of error message
				if tt.exceptedErr != nil {

					r.ErrorIs(err, tt.exceptedErr)
				} else {
					r.Equal(nil, err)
				}
			}

		})
	}
}
