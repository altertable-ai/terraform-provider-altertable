package client

import (
	"context"
	"net/http"
)

func bucketsPath(envID string) string { return "/environments/" + envID + "/buckets" }

// GET /environments/{envID}/buckets
func (c *Client) ListBuckets(ctx context.Context, envID string) ([]Bucket, error) {
	resp, err := request[BucketsListResponse](ctx, c, http.MethodGet, bucketsPath(envID), nil)
	if err != nil {
		return nil, err
	}
	return resp.Buckets, nil
}

// GET /environments/{envID}/buckets/{idOrSlug}
func (c *Client) GetBucket(ctx context.Context, envID, idOrSlug string) (*Bucket, error) {
	resp, err := request[BucketResponse](ctx, c, http.MethodGet, bucketsPath(envID)+"/"+idOrSlug, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// POST /environments/{envID}/buckets
func (c *Client) CreateBucket(ctx context.Context, envID string, params CreateBucketRequest) (*Bucket, error) {
	resp, err := request[BucketResponse](ctx, c, http.MethodPost, bucketsPath(envID), params)
	if err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// PATCH /environments/{envID}/buckets/{id}
func (c *Client) UpdateBucket(ctx context.Context, envID, id string, params UpdateBucketRequest) (*Bucket, error) {
	resp, err := request[BucketResponse](ctx, c, http.MethodPatch, bucketsPath(envID)+"/"+id, params)
	if err != nil {
		return nil, err
	}
	return &resp.Bucket, nil
}

// DELETE /environments/{envID}/buckets/{id}
func (c *Client) DeleteBucket(ctx context.Context, envID, id string) error {
	return c.doRequest(ctx, http.MethodDelete, bucketsPath(envID)+"/"+id, nil, nil)
}
