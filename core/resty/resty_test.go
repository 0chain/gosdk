package resty

import (
	"context"
	"fmt"
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
		expectedErr error
		setup       func(a *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty)
	}{
		{
			name:        "Test_Resty_Cancel_With_Timeout",
			expectedErr: context.DeadlineExceeded,
			urls:        []string{"Test_Resty_Timeout_1", "Test_Resty_Timeout_2", "Test_Resty_Timeout_3"},
			setup:       getSetupFuncForCtxDeadlineExtendedTest(),
		},
		{
			name:        "Test_Resty_All_Success",
			statusCode:  200,
			expectedErr: nil,
			urls:        []string{"http://Test_Resty_Success_1", "http://Test_Resty_Success_2"},
			setup:       getSetupFuncForAllSuccessTest(),
		},
		{
			name:        "Test_Resty_Failed_Due_To_Handler_Failing",
			expectedErr: fmt.Errorf("handler returned error"),
			urls:        []string{"http://Test_Resty_Failure_1", "http://Test_Resty_Failure_2"},
			setup:       getSetupFuncForHandlerErrorTest(),
		},
		{
			name:        "Test_Resty_Failed_Due_To_Interceptor_Error",
			expectedErr: fmt.Errorf("interceptor returned with error"),
			urls:        []string{"http://Test_Resty_Failure_1", "http://Test_Resty_Failure_2"},
			setup:       getSetupFuncForInterceptorErrorTest(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			ctx, resty := tt.setup(r, tt.name, tt.statusCode, tt.urls)
			resty.DoGet(ctx, tt.urls...)
			errs := resty.Wait()
			if tt.expectedErr != nil && tt.expectedErr != context.DeadlineExceeded && len(errs) == 0 {
				t.Fatalf("expected err: %v, got: %v", tt.expectedErr, errs)
			}
			for _, err := range errs {
				// test it by predefined error variable instead of error message
				if tt.expectedErr != nil {
					r.Errorf(err, tt.expectedErr.Error())
				} else {
					r.Equal(nil, err)
				}
			}
		})
	}
}

func getSetupFuncForCtxDeadlineExtendedTest() func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
	return func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
		r := New()
		r.client = &mocks.Timeout{
			Timeout: 1 * time.Second,
		}
		ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return ctx, r
	}
}

func getSetupFuncForAllSuccessTest() func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
	return func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
		resty := New().Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			ra.Equal(200, resp.StatusCode)
			ra.Equal(nil, err)
			ra.Equal(name, string(respBody))
			return nil
		})

		client := &mocks.Client{}
		setupMockClient(client, urls, statusCode, name)
		resty.client = client

		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return context.TODO(), resty
	}
}

func getSetupFuncForHandlerErrorTest() func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
	return func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
		resty := New().Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			return fmt.Errorf("handler returned error")
		})

		client := &mocks.Client{}
		setupMockClient(client, urls, statusCode, name)
		resty.client = client

		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return context.TODO(), resty
	}
}

func getSetupFuncForInterceptorErrorTest() func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
	return func(ra *require.Assertions, name string, statusCode int, urls []string) (context.Context, *Resty) {
		opts := make([]Option, 0)
		opts = append(opts, WithRequestInterceptor(func(r *http.Request) error {
			return fmt.Errorf("interceptor returned with error") //nolint
		}))
		// create a resty object with an interceptor which returns an error, but the handler doesn't return any error
		resty := New(opts...).Then(func(req *http.Request, resp *http.Response, respBody []byte, cf context.CancelFunc, err error) error {
			return nil
		})

		client := &mocks.Client{}
		setupMockClient(client, urls, statusCode, name)
		resty.client = client

		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		go func() {
			<-ctx.Done()
			cancel()
		}()
		return context.TODO(), resty
	}
}

func setupMockClient(mck *mocks.Client, urls []string, statusCode int, name string) {
	for _, url := range urls {
		func(u string) {
			mck.On("Do", mock.MatchedBy(func(r *http.Request) bool {
				return r.URL.String() == u
			})).Return(&http.Response{
				StatusCode: statusCode,
				Body:       ioutil.NopCloser(strings.NewReader(name)),
			}, nil)
		}(url)
	}
}
