package client

import "context"

type GardenClientProvider[T any] func(ctx context.Context, region, key, secret string) (T, error)

type SkrClientProvider[T any] func(ctx context.Context, region, key, secret, role string) (T, error)

type GardenProvider interface {
	Sts() GardenClientProvider[StsClient]
}

type SkrProvider interface {
	Network() SkrClientProvider[NetworkClient]
	Efs() SkrClientProvider[EfsClient]
}

// ==================================================================================

func NewGardenProvider() GardenProvider {
	return &gardenProvider{}
}

type gardenProvider struct{}

func (p *gardenProvider) Sts() GardenClientProvider[StsClient] {
	return StsGardenClientProvider()
}

// ==================================================================================

func NewSkrProvider() SkrProvider {
	return &skrProvider{}
}

type skrProvider struct{}

func (p *skrProvider) Network() SkrClientProvider[NetworkClient] {
	return NetworkSkrProvider()
}

func (p *skrProvider) Efs() SkrClientProvider[EfsClient] {
	return EfsSkrProvider()
}

// ==================================================================================

func StsGardenClientProvider() GardenClientProvider[StsClient] {
	return func(ctx context.Context, region, key, secret string) (StsClient, error) {
		cfg, err := NewConfigBuilder().
			WithRegion(region).
			WithCredentials(key, secret).
			Build(ctx)
		if err != nil {
			return nil, err
		}
		return StsClientFactory(cfg), nil
	}
}

func NetworkSkrProvider() SkrClientProvider[NetworkClient] {
	return func(ctx context.Context, region, key, secret, role string) (NetworkClient, error) {
		cfg, err := NewConfigBuilder().
			WithRegion(region).
			WithCredentials(key, secret).
			WithAssumeRole(role).
			Build(ctx)
		if err != nil {
			return nil, err
		}
		return NetworkClientFactory(cfg), nil
	}
}

func EfsSkrProvider() SkrClientProvider[EfsClient] {
	return func(ctx context.Context, region, key, secret, role string) (EfsClient, error) {
		cfg, err := NewConfigBuilder().
			WithRegion(region).
			WithCredentials(key, secret).
			WithAssumeRole(role).
			Build(ctx)
		if err != nil {
			return nil, err
		}
		return EfsClientFactory(cfg), nil
	}
}
