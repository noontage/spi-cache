package sic

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"reflect"
	mock_local_cache "sp-cache/mock/local_cache"
	mock_redis "sp-cache/mock/redis"
	"testing"
)

func Test_cache_Get(t *testing.T) {
	dsf := func() (interface{}, error) {
		return "BY_DATA_SOURCE", nil
	}
	tests := map[string]struct {
		key      string
		injector func(key string, mr *mock_redis.MockUniversalClient, ml *mock_local_cache.MockLocalCache)
		ds       func() (interface{}, error)
		want     interface{}
		wantErr  bool
	}{
		"キャッシュヒットなし": {
			key: "foo",
			injector: func(key string, mr *mock_redis.MockUniversalClient, ml *mock_local_cache.MockLocalCache) {
				ml.EXPECT().Get(key).Return(nil, false)
				mr.EXPECT().Get(gomock.Any(), key).Return(redis.NewStringResult("", redis.Nil))
			},
			ds:      dsf,
			want:    "BY_DATA_SOURCE",
			wantErr: false,
		},
		"Redis:hit": {
			key: "foo",
			injector: func(key string, mr *mock_redis.MockUniversalClient, ml *mock_local_cache.MockLocalCache) {
				ml.EXPECT().Get(key).Return(nil, false)
				mr.EXPECT().Get(gomock.Any(), key).Return(redis.NewStringResult("HIT_BY_REDIS", nil))
			},
			ds:      dsf,
			want:    "HIT_BY_REDIS",
			wantErr: false,
		},
		"Local:hit": {
			key: "foo",
			injector: func(key string, mr *mock_redis.MockUniversalClient, ml *mock_local_cache.MockLocalCache) {
				ml.EXPECT().Get(key).Return("HIT_BY_LOCAL", true)
			},
			ds:      dsf,
			want:    "HIT_BY_LOCAL",
			wantErr: false,
		},
	}

	for tn, tv := range tests {
		t.Run(tn, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			mr := mock_redis.NewMockUniversalClient(ctrl)
			ml := mock_local_cache.NewMockLocalCache(ctrl)
			tv.injector(tv.key, mr, ml)

			c := NewCache(Config{}, mr, ml)

			got, err := c.Get(ctx, tv.key, tv.ds)
			if (err != nil) != tv.wantErr {
				t.Errorf("wantErr mismatch error=%v, wantErr=%v", err, tv.wantErr)
			}
			if !reflect.DeepEqual(tv.want, got) {
				t.Errorf("want=%v, got=%v", tv.want, got)
			}
		})
	}
}
